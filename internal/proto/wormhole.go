package proto

type WormholeReadWrite struct {
	Wormhole_PenetrateServer
}

func (w *WormholeReadWrite) Read(p []byte) (n int, err error) {
	recv, err := w.Recv()
	if err != nil {
		return 0, err
	}

	n = copy(p, recv.Data)

	return n, nil
}

func (w *WormholeReadWrite) Write(p []byte) (n int, err error) {
	err = w.Send(&PenetrateResponse{
		Data: p,
	})

	return len(p), err
}

type WormholePenetrateClientReadWrite struct {
	Wormhole_PenetrateClient
}

func (w *WormholePenetrateClientReadWrite) Read(p []byte) (n int, err error) {
	recv, err := w.Recv()
	if err != nil {
		return 0, err
	}

	n = copy(p, recv.Data)
	return n, nil
}

func (w *WormholePenetrateClientReadWrite) Write(p []byte) (n int, err error) {
	err = w.Send(&PenetrateRequest{
		Data: p,
	})

	return len(p), err
}
