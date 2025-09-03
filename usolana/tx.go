package usolana


// TxPriorityFee Transaction Priority Fee
// The prioritization fee is an optional fee paid to increase the chance that the current leader processes your transaction.
// https://solana.com/docs/core/fees#prioritization-fee
type TxPriorityFee struct {
	ComputeUnitLimit uint32
	ComputeUnitPrice uint64 // microLamports // 1,000,000 micro-lamports = 1 lamport
}
