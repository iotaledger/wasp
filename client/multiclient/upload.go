package multiclient

import "github.com/iotaledger/wasp/client"

// UploadBlobDataWithQuorum upload data chunks to the blob cache in
// the registries of at least quorum nodes.
func (m *MultiClient) UploadData(fieldValues [][]byte, quorum ...int) error {
	q := m.Len()
	if len(quorum) > 0 {
		q = quorum[0]
	}
	return m.DoWithQuorum(func(i int, client *client.WaspClient) error {
		var err error
		for _, data := range fieldValues {
			_, err = client.PutBlob(data)
			if err != nil {
				return err
			}
		}
		return nil
	}, q)
}
