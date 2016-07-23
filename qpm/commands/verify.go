// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package commands

import (
	"bytes"
	"crypto"
	"flag"
	"fmt"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"io/ioutil"
	"os"
	"path/filepath"
	"qpm.io/common"
	"qpm.io/qpm/core"
	"strings"
)

type VerifyCommand struct {
	BaseCommand
	pkg   *common.PackageWrapper
	fs    *flag.FlagSet
	paths []string
}

func NewVerifyCommand(ctx core.Context) *VerifyCommand {
	return &VerifyCommand{
		BaseCommand: BaseCommand{
			Ctx: ctx,
		},
	}
}

func (v VerifyCommand) Description() string {
	return "Verifies the package PGP signature (experimental)"
}

func (v *VerifyCommand) RegisterFlags(flags *flag.FlagSet) {
	v.fs = flags
}

func (v *VerifyCommand) Run() error {

	var path string
	if v.fs.NArg() > 0 {
		packageName := v.fs.Arg(0)
		path = filepath.Join(core.Vendor, strings.Replace(packageName, ".", string(filepath.Separator), -1))
	} else {
		path = "."
	}

	var err error
	v.pkg, err = common.LoadPackage("")
	if err != nil {
		v.Error(err)
		return err
	}

	// Hash the package

	hash, err := v.hashTree(path)
	if err != nil {
		v.Error(err)
		return err
	}
	fmt.Println("Package SHA-256: " + hash)

	// Verify the signature

	if v.pkg.Version.Fingerprint == "" {
		err = fmt.Errorf("no fingerprint set in " + core.PackageFile)
		v.Error(err)
		return err
	}

	entity, err := entityFromLocal("pubring.gpg", v.pkg.Version.Fingerprint)
	if err != nil {
		v.Error(err)
		return err
	}

	sig, err := ioutil.ReadFile(core.SignatureFile)
	if err != nil {
		v.Error(err)
		return err
	}

	err = Verify(hash, sig, entity.PrimaryKey)
	if err != nil {
		v.Error(err)
		return err
	}

	fmt.Println("Signature verified")

	return nil
}

func (v *VerifyCommand) visit(path string, f os.FileInfo, err error) error {

	if f.IsDir() {
		if strings.HasPrefix(f.Name(), ".git") {
			return filepath.SkipDir
		}
	} else {
		v.paths = append(v.paths, path)
	}
	return nil
}

func (v *VerifyCommand) hashTree(directory string) (string, error) {

	v.paths = []string{}

	err := filepath.Walk(directory, v.visit)
	if err != nil {
		return "", err
	}

	return HashPaths(v.paths)
}

func Verify(payload string, signature []byte, pubkey *packet.PublicKey) error {

	// decode and read the signature

	block, err := armor.Decode(bytes.NewBuffer(signature))
	if err != nil {
		return err
	}

	pkt, err := packet.Read(block.Body)
	if err != nil {
		return err
	}

	sig, ok := pkt.(*packet.Signature)
	if !ok {
		return fmt.Errorf("could not parse the signature")
	}

	if sig.Hash != crypto.SHA256 || sig.Hash != crypto.SHA512 {
		return fmt.Errorf("was not a SHA-256 or SHA-512 signature")
	}

	if sig.SigType != packet.SigTypeBinary {
		return fmt.Errorf("was not a binary signature")
	}

	// verify the signature

	hash := sig.Hash.New()
	_, err = hash.Write([]byte(payload))
	if err != nil {
		return err
	}

	err = pubkey.VerifySignature(hash, sig)
	if err != nil {
		return err
	}

	return nil
}
