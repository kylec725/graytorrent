package main


func newTorrent(filename string, port uint16) (torrent.Torrent, error) {
    to := torrent.Torrent{Path: filename, Port: port}
    if err := to.Setup(); err != nil {
        log.WithFields(log.Fields{"file": filename, "error": err.Error()}).Info("Torrent setup failed")
        return torrent.Torrent{}, err
    }
    torrentList = append(torrentList, to)
    log.WithField("name", to.Info.Name).Info("Torrent added")
    return to, nil
}
