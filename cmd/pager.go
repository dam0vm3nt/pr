package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pterm/pterm"
	"github.com/vballestra/gobb-cli/bitbucket"
	"io/ioutil"
	"net/http"
)

type Paginated[T any, C any] interface {
	GetContainer() *C
	GetNext() string
	GetPages() int32
	GetValues() []T
}

func Paginate[T any, C any](ctx context.Context, pager Paginated[T, C]) <-chan T {
	container := pager.GetContainer()

	c := make(chan T)

	// Fetch pages
	spinner := pterm.DefaultSpinner.WithRemoveWhenDone(true)
	spinner.Start()

	go func() {

		count := int32(0)

		for cont := true; cont && count < pager.GetPages(); {
			for _, val := range pager.GetValues() {
				c <- val
				count += 1
			}
			spinner.UpdateText(fmt.Sprintf("Loaded page %d/%d", count, pager.GetPages()))

			if len(pager.GetNext()) > 0 {
				req, _ := http.NewRequestWithContext(ctx, "GET", pager.GetNext(), nil)
				if auth, ok := ctx.Value(bitbucket.ContextBasicAuth).(bitbucket.BasicAuth); ok {
					req.SetBasicAuth(auth.UserName, auth.Password)
				}

				resp2, _ := http.DefaultClient.Do(req)
				bb, _ := ioutil.ReadAll(resp2.Body)

				err := json.Unmarshal(bb, container)
				if err != nil {
					spinner.Fail("Error!")
					return // pterm.Fatal.Println(err)
				}
				cont = true
			} else {
				cont = false
			}
		}
		spinner.Success("Finished")
		close(c)
	}()

	return c
}
