package internal

import (
	"time"

	"github.com/yarlson/tap"
)

// ShowProgress displays a progress bar for long-running operations
func ShowProgress(message, successMessage string, operation func() error) error {
	prog := tap.NewProgress(tap.ProgressOptions{
		Style: "heavy",
		Max:   100,
		Size:  40,
	})

	prog.Start(message)

	// Run operation in background
	done := make(chan error, 1)
	go func() {
		done <- operation()
	}()

	// Animate progress while operation runs
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	progress := 0
	for {
		select {
		case err := <-done:
			prog.Advance(100-progress, "Complete")
			if err != nil {
				prog.Stop("Failed", 2)
				return err
			} else {
				prog.Stop(successMessage, 0)
			}
			return nil
		case <-ticker.C:
			if progress < 90 {
				advancement := 5 + (progress / 10) // Slow down as we get closer
				prog.Advance(advancement, message)
				progress += advancement
			}
		}
	}
}
