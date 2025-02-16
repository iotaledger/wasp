package l1simulator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const DefaultHost = "localhost:30003"

func TestNewClient(t *testing.T) {
	r := require.New(t)
	client := NewClient("http://" + DefaultHost)

	r.Equal("http://"+DefaultHost, client.baseURL)
	r.NotNil(client.httpClient)
}

func TestHealth(t *testing.T) {
	r := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	health, err := client.Health()
	r.NoError(err)
	r.Equal("OK", health)
}

func TestCreateCheckpoint(t *testing.T) {
	r := require.New(t)

	expectedCheckpoint := &Checkpoint{
		Summary: CheckpointSummary{
			Epoch:                    1,
			SequenceNumber:           100,
			NetworkTotalTransactions: 1000,
			ContentDigest:            "digest123",
			TimestampMs:              1234567890,
			CheckpointCommitments:    []string{"commit1", "commit2"},
			VersionSpecificData:      []byte("data"),
		},
		Authority: AuthorityStrongQuorumSignInfo{
			Epoch:      1,
			Signature:  "sig123",
			SignersMap: make([]uint8, 0),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/create_checkpoint", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedCheckpoint)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	checkpoint, err := client.CreateCheckpoint()
	r.NoError(err)
	r.Equal(expectedCheckpoint.Summary.Epoch, checkpoint.Summary.Epoch)
}

func TestAdvanceClock(t *testing.T) {
	r := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/advance_clock", r.URL.Path)

		var req AdvanceClockRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, uint32(1000), req.Duration)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.AdvanceClock(1000)
	r.NoError(err)
}

func TestAdvanceEpoch(t *testing.T) {
	r := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/advance_epoch", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.AdvanceEpoch()
	r.NoError(err)
}

func TestErrorHandling(t *testing.T) {
	r := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.CreateCheckpoint()
	r.Error(err)

	err = client.AdvanceClock(1000)
	r.Error(err)

	err = client.AdvanceEpoch()
	r.Error(err)
}

func TestInvalidURL(t *testing.T) {
	r := require.New(t)
	client := NewClient("http://invalid-url-that-does-not-exist:12345")

	_, err := client.Health()
	r.Error(err)

	_, err = client.CreateCheckpoint()
	r.Error(err)

	err = client.AdvanceClock(1000)
	r.Error(err)

	err = client.AdvanceEpoch()
	r.Error(err)
}

const IntegrationTestEnv = "RUN_INTEGRATION_TESTS"

func TestIntegration(t *testing.T) {
	if os.Getenv(IntegrationTestEnv) != "1" {
		t.Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=1 to enable")
	}

	r := require.New(t)
	baseURL := fmt.Sprintf("http://%s", DefaultHost)
	err := waitForServer(baseURL)
	r.NoError(err)

	client := NewClient(baseURL)

	t.Run("IntegrationHealth", func(t *testing.T) {
		r := require.New(t)
		health, err := client.Health()
		r.NoError(err)
		r.Equal("OK", health)
	})

	t.Run("IntegrationCreateCheckpoint", func(t *testing.T) {
		r := require.New(t)
		checkpoint, err := client.CreateCheckpoint()
		r.NoError(err)
		r.NotNil(checkpoint)
	})

	t.Run("IntegrationAdvanceClock", func(t *testing.T) {
		r := require.New(t)
		err := client.AdvanceClock(1000)
		r.NoError(err)
	})

	t.Run("IntegrationAdvanceEpoch", func(t *testing.T) {
		r := require.New(t)
		err := client.AdvanceEpoch()
		r.NoError(err)
	})

	t.Run("IntegrationFullFlow", func(t *testing.T) {
		r := require.New(t)

		checkpoint1, err := client.CreateCheckpoint()
		r.NoError(err)

		err = client.AdvanceClock(2000)
		r.NoError(err)

		checkpoint2, err := client.CreateCheckpoint()
		r.NoError(err)
		r.Greater(checkpoint2.Summary.TimestampMs, checkpoint1.Summary.TimestampMs)

		err = client.AdvanceEpoch()
		r.NoError(err)

		checkpoint3, err := client.CreateCheckpoint()
		r.NoError(err)
		r.Greater(checkpoint3.Summary.Epoch, checkpoint2.Summary.Epoch)
	})
}

func waitForServer(baseURL string) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("server failed to respond within timeout")
}
