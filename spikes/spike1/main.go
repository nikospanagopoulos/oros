package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func main() {
	// Το πρόγραμμα-στόχος δίνεται ως argument (π.χ. ./wrapper python3),
	// με default το bash αν δεν δοθεί τίποτα.
	target := "bash"
	var targetArgs []string
	if len(os.Args) > 1 {
		target = os.Args[1]
		targetArgs = os.Args[2:]
	}
	cmd := exec.Command(target, targetArgs...)

	// pty.Start() κάνει τρία πράγματα ταυτόχρονα:
	// 1. Δημιουργεί ένα ζευγάρι PTY (master + slave)
	// 2. Συνδέει το slave άκρο ως Stdin/Stdout/Stderr του child process
	// 3. Ξεκινάει το child process
	// Μας επιστρέφει το master άκρο, σαν ένα *os.File που μπορούμε να
	// διαβάζουμε (Read) και να γράφουμε (Write).
	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Αποτυχία εκκίνησης PTY: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = ptmx.Close() }()

	fmt.Println("=== [oros-spike1] Ξεκινάει το child process (bash) μέσω PTY ===")

	// Goroutine: ό,τι πληκτρολογεί ο πραγματικός χρήστης (os.Stdin)
	// το αντιγράφουμε συνεχώς προς το PTY master -> φτάνει στο child.
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()

	// Στο κύριο thread: ό,τι τυπώνει το child στο PTY master (δηλαδή
	// stdout/stderr του bash) το αντιγράφουμε προς το πραγματικό
	// terminal του χρήστη (os.Stdout). Μπλοκάρει μέχρι να κλείσει το PTY.
	_, _ = io.Copy(os.Stdout, ptmx)

	err = cmd.Wait()
	if err != nil {
		fmt.Printf("\n=== [oros-spike1] Το child τελείωσε με σφάλμα: %v ===\n", err)
	} else {
		fmt.Println("\n=== [oros-spike1] Το child τελείωσε κανονικά (exit 0) ===")
	}
}
