package proto

type WormholeReadWrite struct {
	Wormhole_PenetrateServer
}

func (w *WormholeReadWrite) Read(p []byte) (n int, err error) {
	recv, err := w.Recv()
	if err != nil {
		return 0, err
	}

	p = recv.Data
	return len(recv.Data), nil
}

func (w *WormholeReadWrite) Write(p []byte) (n int, err error) {
	err = w.Send(&PenetrateResponse{
		Data: p,
	})

	return len(p), err
}
