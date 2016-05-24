// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
	"qpm.io/common"
	msg "qpm.io/common/messages"
	"qpm.io/qpm/core"
	"qpm.io/qpm/vcs"
)

type SignCommand struct {
	BaseCommand
	pkg   *common.PackageWrapper
	paths []string
}

func NewSignCommand(ctx core.Context) *SignCommand {
	return &SignCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (s SignCommand) Description() string {
	return "Creates a PGP signature for the package (experimental)"
}

func (s *SignCommand) RegisterFlags(flags *flag.FlagSet) {

}

func (s *SignCommand) Run() error {

	var err error
	s.pkg, err = common.LoadPackage("")
	if err != nil {
		s.Error(err)
		return err
	}

	// Hash the repo

	hash, err := hashRepo(s.pkg.Repository)
	if err != nil {
		s.Error(err)
		return err
	}
	fmt.Println("Package SHA-256: " + hash)

	// Sign the SHA

	fmt.Println("Loading the GnuPG private key")

	if s.pkg.Version.Fingerprint == "" {
		err = fmt.Errorf("no fingerprint set in " + core.PackageFile)
		s.Error(err)
		return err
	}

	signer, err := entityFromLocal("secring.gpg", s.pkg.Version.Fingerprint)
	if err != nil {
		s.Error(err)
		return err
	}

	fmt.Println("Creating the signature")
	sig, err := Sign(hash, signer)
	if err != nil {
		s.Error(err)
		return err
	}

	// Write the signature file

	fmt.Println("Creating " + core.SignatureFile)
	file, err := os.Create(core.SignatureFile)
	if err != nil {
		s.Error(err)
		return err
	}
	defer file.Close()

	_, err = file.Write(sig)
	if err != nil {
		s.Error(err)
		return err
	}

	// Verify the signature

	fmt.Println("Verifying the signature")
	entity, err := entityFromLocal("pubring.gpg", s.pkg.Version.Fingerprint)
	if err != nil {
		s.Error(err)
		return err
	}

	err = Verify(hash, sig, entity.PrimaryKey)
	if err != nil {
		s.Error(err)
		return err
	}

	fmt.Println("Done")

	return nil
}

// SHA-256 hashing

func hash(path string) ([]byte, error) {

	var result []byte
	file, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}

func HashPaths(paths []string) (string, error) {

	var result []byte

	// we need to sort to get consistent results
	sort.Strings(paths)

	master := sha256.New()

	for _, p := range paths {
		f, err := os.Stat(p)
		if err != nil {
			return "", err
		}
		if strings.HasSuffix(p, core.SignatureFile) {
			continue
		}
		if !f.IsDir() {
			sha, err := hash(p)
			if err != nil {
				fmt.Println(err)
				return "", err
			}
			master.Write(sha)
		}
	}

	result = master.Sum(nil)

	return hex.EncodeToString(result), nil
}

func hashRepo(repository *msg.Package_Repository) (string, error) {

	publisher, err := vcs.CreatePublisher(repository)
	if err != nil {
		return "", err
	}

	paths, err := publisher.RepositoryFileList()
	if err != nil {
		return "", err
	}

	return HashPaths(paths)
}

// PGP signing

func decryptEntity(entity *openpgp.Entity) error {

	passwd := <-PromptPassword("Password:")
	err := entity.PrivateKey.Decrypt([]byte(passwd))
	if err == nil {
		return nil
	}
	return nil
}

func entityFromLocal(fileName string, fingerprint string) (*openpgp.Entity, error) {

	path := os.Getenv("GNUPGHOME")
	if len(path) == 0 {
		return nil, fmt.Errorf("cound not find GNUPGHOME in ENV")
	}

	file, err := os.Open(filepath.Join(path, fileName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	keyring, err := openpgp.ReadKeyRing(file)
	if err != nil {
		return nil, err
	}

	decoded, err := hex.DecodeString(fingerprint)
	if err != nil {
		return nil, err
	}

	if len(decoded) != 20 {
		return nil, fmt.Errorf("the fingerprint is not 20 bytes")
	}

	var fp [20]byte
	copy(fp[:], decoded[:20])

	for _, entity := range keyring {

		if entity.PrimaryKey.Fingerprint != fp {
			continue
		}

		if entity != nil && entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
			if err := decryptEntity(entity); err != nil {
				return nil, err
			}
			return entity, nil
		}
		return entity, nil
	}

	return nil, fmt.Errorf("entity for %s not found in %s", fingerprint, fileName)
}

func Sign(unsigned string, signer *openpgp.Entity) ([]byte, error) {

	var buffer bytes.Buffer
	err := openpgp.ArmoredDetachSign(&buffer, signer, strings.NewReader(unsigned), &packet.Config{})
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
