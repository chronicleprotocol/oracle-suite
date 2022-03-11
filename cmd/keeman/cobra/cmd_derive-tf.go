//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cobra

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/spf13/cobra"
)

type query struct {
	Mnemonic string `json:"mnemonic"`
	Path     string `json:"path"`
	Password string `json:"password"`
	Format   string `json:"format"`
}

func NewDeriveTf() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "derive-tf",
		Short: "Derive keys from HD Mnemonic (Terraform External Data style)",
		RunE: func(_ *cobra.Command, _ []string) error {
			var q query
			err := json.NewDecoder(os.Stdin).Decode(&q)
			if err != nil {
				return fmt.Errorf("input decoding failed: %w", err)
			}
			wallet, err := hdwallet.NewFromMnemonic(q.Mnemonic)
			if err != nil {
				return err
			}
			dp, err := accounts.ParseDerivationPath(q.Path)
			if err != nil {
				return err
			}
			acc, err := wallet.Derive(dp, false)
			if err != nil {
				return err
			}
			privateKey, err := wallet.PrivateKey(acc)
			if err != nil {
				return err
			}
			b, err := formattedBytes(q.Format, privateKey, q.Password)
			if err != nil {
				return err
			}
			fmt.Printf(
				`{"output":"%s","path":"%s","addr":"%s"}`,
				base64.StdEncoding.EncodeToString(b),
				dp.String(),
				acc.Address.String(),
			)
			return nil
		},
	}
	return cmd
}
