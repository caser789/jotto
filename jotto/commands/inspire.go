package commands

import (
	"flag"
	"fmt"

	"git.garena.com/duanzy/motto/motto"
)

type Quote struct {
	quote string
	name  string
}

// Inspire is a demo command
type Inspire struct {
	motto.BaseCommand
	number *int
	quotes []*Quote
	cmd    *flag.FlagSet
}

func NewInspire() *Inspire {
	return &Inspire{
		quotes: []*Quote{
			&Quote{"The best preparation for tomorrow is doing your best today.", "H. Jackson Brown, Jr."},
			&Quote{"It is during our darkest moments that we must focus to see the light.", "Aristotle"},
			&Quote{"We must let go of the life we have planned, so as to accept the one that is waiting for us.", "Joseph Campbell"},
			&Quote{"Put your heart, mind, and soul into even your smallest acts. This is the secret of success.", "Swami Sivananda"},
			&Quote{"Try to be rainbow in someone's cloud.", "Maya Angelow"},
			&Quote{"Change your thoughts and you changes your world.", "Norman Vincent Peale"},
			&Quote{"No act of kindness, no matter how small, is ever wasted.", "Aesop"},
			&Quote{"If opportunity doesn't knock, build a door.", "Milton Berle"},
			&Quote{"We know what we are, but know not what we may be.", "William Shakespeare"},
			&Quote{"What we think, we become.", "Buddha"},
		},
	}
}

func (i *Inspire) Name() string {
	return "demo:inspire"
}

func (i *Inspire) Description() string {
	return "Inspire is a command for inspiration."
}

func (i *Inspire) Boot() (err error) {
	i.cmd = flag.NewFlagSet(i.Name(), flag.ExitOnError)
	i.number = i.cmd.Int("number", 1, "Which inspirational quote to display (1-10).")

	return
}

func (i *Inspire) Run(app motto.Application, args []string) (err error) {

	quote := i.quotes[*i.number-1]

	fmt.Printf("%s\n    -- %s\n", quote.quote, quote.name)

	// Example: Access the application instance.
	fmt.Printf("Protocol: %s\n", app.Protocol())

	return
}
