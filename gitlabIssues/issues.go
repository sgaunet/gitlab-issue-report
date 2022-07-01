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
		fmt.Printf("%-70s %10s %-25s %-25s\n", "Title", "State", "Created At", "Updated At")
	}
	for idx := range i {
		fmt.Printf("%-70s %10s %25s %25s\n", truncateStr(i[idx].Title, 70), i[idx].State, i[idx].CreatedAt.Format("2006-01-02T15:04:05-0700"), i[idx].UpdateAt.Format("2006-01-02T15:04:05-0700"))
	}
}

func truncateStr(str string, length int) string {
	if len(str) > length && length > 0 {
		return str[:length]
	}
	return str
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
