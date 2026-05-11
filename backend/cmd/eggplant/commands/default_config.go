package commands

import (
	"os"

	"github.com/boreq/eggplant/internal/config"
	"github.com/boreq/errors"
	"github.com/boreq/guinea"
)

var defaultConfigCmd = guinea.Command{
	Run:              runDefaultConfig,
	ShortDescription: "prints default configuration to stdout",
}

func runDefaultConfig(c guinea.Context) error {
	conf := config.Default()

	if err := config.Marshal(os.Stdout, conf.ExposedConfig); err != nil {
		return errors.Wrap(err, "failed to marshal the default config")
	}

	return nil
}
