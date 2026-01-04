package iotagraphql

import client "github.com/iotaledger/wasp/v2/clients/iota-go/client"

var _ client.IotaClient = (*GraphQLClient)(nil)
