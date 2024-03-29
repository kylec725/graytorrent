package cmd

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/kylec725/graytorrent/internal/config"
	pb "github.com/kylec725/graytorrent/rpc"
	"github.com/kylec725/graytorrent/torrent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const pidFile = "/tmp/graytorrent.pid"

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverMainCmd)
	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStopCmd)
	serverMainCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "logs additional information for debugging")
	serverStartCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "logs additional information for debugging")
}

var (
	server *grpc.Server

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "controls the graytorrent server",
	}

	serverMainCmd = &cobra.Command{
		Use:    "main",
		Hidden: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			initLog()
		},
		Run: func(cmd *cobra.Command, args []string) {
			session, err := torrent.NewSession()
			if err != nil {
				log.WithField("error", err).Info("Error when starting a new session for server")
			}

			// Initialize signal catching
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)
			go func() {
				_ = <-signalChan
				signal.Stop(signalChan)

				// Cleanup
				session.Close()
				server.Stop()
				logFile.Close()

				// Remove PID file
				os.Remove(pidFile)

				os.Exit(0)

			}()

			// Setup grpc server
			// TODO: Want to use TLS for encrypting communication
			serverAddr := ":" + strconv.Itoa(config.GetConfig().Network.ServerPort)
			serverListener, err := net.Listen("tcp", serverAddr)
			if err != nil {
				log.WithFields(log.Fields{"error": err.Error(), "port": serverAddr[1:]}).Fatal("Failed to listen for rpc")
			}
			server = grpc.NewServer()
			pb.RegisterTorrentServiceServer(server, &session)
			if err = server.Serve(serverListener); err != nil {
				log.WithField("error", err).Debug("Error with serving rpc client")
			}
		},
	}

	serverStartCmd = &cobra.Command{
		Use:   "start",
		Short: "starts the graytorrent server",
		Run: func(cmd *cobra.Command, args []string) {
			// Check if daemon already running.
			if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
				data, err := ioutil.ReadFile(pidFile)
				if err != nil {
					fmt.Println("Unable to read process ID found in ", pidFile)
					os.Exit(1)
				}

				pid, err := strconv.Atoi(string(data))
				if err != nil {
					fmt.Println("Unable to parse process ID found in ", pidFile)
					os.Exit(1)
				}

				process, err := os.FindProcess(pid)
				if err := process.Signal(syscall.Signal(0)); err == nil {
					fmt.Println("graytorrent is already running")
					os.Exit(1)
				}
			}

			// Use string slice to forward flags
			serverMain := []string{serverCmd.Use, serverMainCmd.Use}
			if debug {
				serverMain = append(serverMain, "-d")
			}

			daemon := exec.Command(os.Args[0], serverMain...)
			daemon.Start()
			savePID(daemon.Process.Pid)

			// fmt.Println("Daemon process ID is : ", daemon.Process.Pid)
			fmt.Println("graytorrent started")
			os.Exit(0)

		},
	}

	serverStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "stops the graytorrent server",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat(pidFile); os.IsNotExist(err) { // do nothing if pid file does not exist
				fmt.Println("graytorrent is not running")
				os.Exit(1)
			}

			data, err := ioutil.ReadFile(pidFile)
			if err != nil {
				fmt.Println("Unable to read process ID found in ", pidFile)
				os.Exit(1)
			}

			pid, err := strconv.Atoi(string(data))
			if err != nil {
				fmt.Println("Unable to parse process ID found in ", pidFile)
				os.Exit(1)
			}

			// remove PID file
			os.Remove(pidFile)

			if err = syscall.Kill(pid, syscall.SIGTERM); err != nil {
				fmt.Printf("Unable to kill process ID [%v] with error %v \n", pid, err) // Change to graytorrent is not running in the future
				os.Exit(1)
			}

			// kill process and exit immediately
			// if err = process.Signal(os.Interrupt); err != nil {
			// 	fmt.Printf("Unable to kill process ID [%v] with error %v \n", pid, err)
			// 	os.Exit(1)
			// }

			fmt.Println("graytorrent stopped")
			os.Exit(0)
		},
	}
)

func savePID(pid int) {
	file, err := os.Create(pidFile)
	if err != nil {
		log.Fatalf("Unable to create pid file : %v\n", err)
	}
	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))
	if err != nil {
		log.Fatalf("Unable to create pid file : %v\n", err)
	}
	file.Sync() // Flush to disk
}
