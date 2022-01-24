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
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
)

func GenerateSeed(opts *Options) *cobra.Command {
	var bitSizeMultiplier int
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen", "g"},
		Args:    cobra.NoArgs,
		Short:   "Generate HD seed phrase with a specific bit size",
		RunE: func(_ *cobra.Command, args []string) error {
			if bitSizeMultiplier < 4 || bitSizeMultiplier > 8 {
				return fmt.Errorf("entropy size multiplier must be between 4 and 8")
			}
			bitSize := 32 * bitSizeMultiplier
			log.Printf("entropy bit size: %d * 32 = %d\n", bitSizeMultiplier, bitSize)
			entropy, err := bip39.NewEntropy(bitSize)
			if err != nil {
				return err
			}
			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				return err
			}
			fmt.Println(mnemonic)
			if opts.Verbose {
				log.Println(mnemonic)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(
		&bitSizeMultiplier,
		"entropy",
		"n",
		4,
		"number of 32 bit size blocks for entropy <4;8>",
	)
	return cmd
}
