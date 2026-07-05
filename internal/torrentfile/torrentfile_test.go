package torrentfile

import (
	"encoding/hex"
	"testing"
)

func TestOpen(t *testing.T) {
	tf, err := Open("testdata/debian-13.5.0-amd64-netinst.iso.torrent")
	if err != nil {
		t.Fatalf("unnable to open file: %v", err)
	}

	got := hex.EncodeToString(tf.InfoHash[:])
	if got != "58846860f0a766f8a42b0bb214d8c713fdf1b167" {
		t.Errorf("info_hash = %s, want 58846860f0a766f8a42b0bb214d8c713fdf1b167", got)
	}

	if tf.Announce != "http://bttracker.debian.org:6969/announce" {
		t.Errorf("announce = %s, want http://bttracker.debian.org:6969/announce", tf.Announce)
	}

	if tf.Name != "debian-13.5.0-amd64-netinst.iso" {
		t.Errorf("Name = %s, want debian-13.5.0-amd64-netinst.iso", tf.Name)
	}

	if tf.Length != 791674880 {
		t.Errorf("Length = %v, 791674880 got ", tf.Length)
	}

	if tf.PieceLength != 262144 {
		t.Errorf("Piece_length = %v, want 262144  got", tf.PieceLength)
	}

	if len(tf.PieceHashes) != 3020 {
		t.Errorf("PieceHashes total = %v, want  3020", len(tf.PieceHashes))
	}
}
