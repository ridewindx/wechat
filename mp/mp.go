package mp

type MP struct {
	*Server
	*Client
	CorpClient *Client
}

func New(id, appID, appSecret, token, aesKey string, urlPrefix ...string) *MP {
	client := NewClient(appID, appSecret, true)
	server := NewServer(token, aesKey, urlPrefix...)
	server.SetClient(client)
	server.SetID(id)
	server.SetAppID(appID)
	return &MP{
		Server: server,
		Client: client,
	}
}
