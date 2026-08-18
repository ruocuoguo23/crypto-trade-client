package gagliardetto

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
)

func TestSystemTransferInstruction(t *testing.T) {
	from := solana.NewWallet().PublicKey()
	to := solana.NewWallet().PublicKey()
	transfer := system.NewTransferInstruction(12345, from, to)

	if transfer.Lamports == nil || *transfer.Lamports != 12345 {
		t.Fatalf("lamports = %v", transfer.Lamports)
	}
	if transfer.GetFundingAccount().PublicKey != from {
		t.Fatalf("funding account = %s, want %s", transfer.GetFundingAccount().PublicKey, from)
	}
	if transfer.GetRecipientAccount().PublicKey != to {
		t.Fatalf("recipient account = %s, want %s", transfer.GetRecipientAccount().PublicKey, to)
	}
}
