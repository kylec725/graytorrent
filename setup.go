package main

// func catchInterrupt(ctx context.Context, cancel context.CancelFunc) {
// 	signalChan := make(chan os.Signal, 1)
// 	signal.Notify(signalChan, os.Interrupt)
// 	select {
// 	case <-signalChan: // Cleanup on interrupt signal
// 		signal.Stop(signalChan)
// 		peerListener.Close()
// 		cancel()
// 		err = torrent.SaveAll(torrentList)
// 		if err != nil {
// 			log.WithField("error", err).Debug("Problem occurred while saving torrent management data")
// 		}
// 		log.Info("Graytorrent stopped")
// 		logFile.Close()
// 		os.Exit(1)
// 	case <-ctx.Done():
// 	}
// }
