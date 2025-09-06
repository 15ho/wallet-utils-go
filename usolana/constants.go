package usolana

const (
	LamportsPerSOL          uint64 = 1000000000                               // 1 SOL = 1,000,000,000 Lamports
	MicroLamportsPerLamport uint64 = 1000000                                  // 1 Lamport = 1,000,000 MicroLamports
	MicroLamportsPerSOL     uint64 = MicroLamportsPerLamport * LamportsPerSOL // 1 SOL = 1,000,000,000,000 MicroLamports
	SOLDecimals             uint8  = 9
)
