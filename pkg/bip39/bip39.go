/*
Package bip39 implements the BIP-0039 specification of using a mnemonic sentence
to encode/decode binary entropy and to derive a seed that can then be used to
generate deterministic wallets.

This package supports entropy lengths from 128 to 512 bits as long as they are a
multiple of 32 bits.

It comes with the official English and Japanese word list, but different word
lists can be registered using RegisterWordList as long as they fullfil the
requirements for a BIP-0039 word list.

This package is tested against the test vectors provided in the official
BIP-0039 specification.
*/
package bip39

import (
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"log"
	"math/big"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/text/unicode/norm"

	"github.com/iotaledger/iota-crypto-demo/pkg/bip39/internal/wordlists"
	"github.com/iotaledger/iota-crypto-demo/pkg/bip39/wordlist"
)

var (
	// ErrInvalidEntropySize is returned when trying to use an entropy with an invalid size.
	ErrInvalidEntropySize = errors.New("invalid entropy size")
	// ErrInvalidMnemonic is returned when trying to use a malformed mnemonic.
	ErrInvalidMnemonic = errors.New("invalid mnemonic")
	// ErrInvalidChecksum is returned when checksum does not match.
	ErrInvalidChecksum = errors.New("invalid checksum")
)

const (
	// SeedSize is the size, in bytes, of a BIP-39 seed.
	SeedSize = 64

	// default word list language.
	defaultLanguage = "english"
)

func init() {
	// register internal word lists
	RegisterWordList("english", wordlists.English)
	RegisterWordList("japanese", wordlists.Japanese)

	// enable default language
	if err := SetWordList(defaultLanguage); err != nil {
		log.Fatalf("error setting default language: %s", err)
	}
}

var wordList wordlist.List

// MnemonicToSeed creates a hashed seed output given a provided string and password.
// No checking is performed to validate that the string provided is a valid mnemonic.
func MnemonicToSeed(mnemonic Mnemonic, passphrase string) ([]byte, error) {
	// validate mnemonic
	if _, err := MnemonicToEntropy(mnemonic); err != nil {
		return nil, err
	}
	// UTF-8 NFKD
	passphrase = norm.NFKD.String(passphrase)
	key := pbkdf2.Key([]byte(mnemonic.String()), []byte("mnemonic"+passphrase), 2048, SeedSize, sha512.New)
	return key, nil
}

// EntropyToMnemonic generates a BIP-39 mnemonic sentence that satisfies the given entropy length.
func EntropyToMnemonic(entropy []byte) (Mnemonic, error) {
	if err := validateEntropy(entropy); err != nil {
		return nil, err
	}

	// compute entropy bit count, denoted by ENT
	bitsEntropy := len(entropy) * 8

	// the checksum is generated by taking the first ENT / 32 bits of the entropy's SHA256 hash
	bitsChecksum := bitsEntropy / 32
	checksum := computeChecksum(entropy, bitsChecksum)

	// the checksum is appended to the end of the initial entropy
	bigEntropy := new(big.Int).SetBytes(entropy)
	bigEntropy.Lsh(bigEntropy, uint(bitsChecksum))
	bigEntropy.Or(bigEntropy, checksum)

	// allocate the number of mnemonic words soon to be generated
	words := make(Mnemonic, entropyBitsToWordCount(bitsEntropy))

	// split into groups of 11 bits, each encoding a number from 0-2047, serving as an index into a word list
	wordIndex := big.NewInt(0)
	for i := len(words) - 1; i >= 0; i-- {
		// get least significant 11 bits
		wordIndex.And(bigEntropy, wordIndexMask)
		// convert the 11 bits to index
		words[i] = wordList.Word(int(wordIndex.Int64()))

		// shift out least significant 11 bits
		bigEntropy.Rsh(bigEntropy, wordlist.IndexBits)
	}

	return words, nil
}

// MnemonicToEntropy takes a BIP-39 mnemonic sentence and returns the initial
// entropy used. If the sentence is invalid, an error is returned.
func MnemonicToEntropy(mnemonic Mnemonic) ([]byte, error) {
	if err := validateMnemonic(mnemonic); err != nil {
		return nil, err
	}

	// compute bit counts
	bitsEntropy := wordCountToEntropyBits(len(mnemonic))
	bitsChecksum := bitsEntropy / entropyMultiple

	// use a big.Int to decode words for easier bitwise operations
	decoder := big.NewInt(0)
	for _, word := range mnemonic {
		wordIndex := wordList.Index(word)
		if wordIndex < 0 || wordIndex >= wordlist.Count {
			panic("invalid word index")
		}

		decoder.Lsh(decoder, wordlist.IndexBits)
		decoder.Or(decoder, big.NewInt(int64(wordIndex)))
	}

	// the checksum corresponds to the last few bits of the decoded bytes
	checksumMask := new(big.Int).Lsh(bigOne, uint(bitsChecksum))
	checksumMask.Sub(checksumMask, bigOne)
	checksum := new(big.Int).And(decoder, checksumMask)

	entropy := decoder.Rsh(decoder, uint(bitsChecksum)).Bytes()
	entropy = padBytes(entropy, bitsEntropy/8)

	// check whether the decoded checksum matches the computed
	if checksum.Cmp(computeChecksum(entropy, bitsChecksum)) != 0 {
		return nil, ErrInvalidChecksum
	}
	return entropy, nil
}

// computeChecksum computes the checksum of the given bytes by returning the first numBits of the SHA256 hash.
func computeChecksum(bytes []byte, numBits int) *big.Int {
	const bitsHash = sha256.Size * 8
	if numBits > bitsHash {
		panic("invalid number of bits")
	}

	// compute hash of bytes
	hash := sha256.Sum256(bytes)

	// take the first numBits of the hash
	checksum := new(big.Int).SetBytes(hash[:])
	return checksum.Rsh(checksum, uint(bitsHash-numBits))
}
