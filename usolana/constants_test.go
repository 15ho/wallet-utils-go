package usolana

import "os"

var (
	Acc1PrivateKeyBase58 = os.Getenv("ACC1PK58")
	Acc1AccountAddress   = os.Getenv("ACC1ADDR")

	Acc2PrivateKeyBase58 = os.Getenv("ACC2PK58")
	Acc2AccountAddress   = os.Getenv("ACC2ADDR")
)
