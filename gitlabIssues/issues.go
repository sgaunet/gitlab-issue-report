package gitlabissues

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
)

type Issues []Issue

func (i Issues) PrintOneLine(printHeader bool) {
	if printHeader {
		fmt.Printf("%50s %10s %s", "Title", "State", "Updated At")
	}
	for idx := range i {
		fmt.Println(i[idx].Title)
		fmt.Println(i[idx].State)
		fmt.Println(i[idx].UpdateAt)
		fmt.Printf("%50s %10s %s", i[idx].Title, i[idx].State, i[idx].UpdateAt)
	}
}

func (i Issues) PrintTab() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Title", "State", "CreatedAt", "UpdatedAt"})

	for _, v := range i {
		t := []string{v.Title, v.State, v.CreatedAt.Format(time.RFC3339), v.UpdateAt.Format(time.RFC3339)}
		table.Append(t)
	}
	table.Render() // Send output
}
