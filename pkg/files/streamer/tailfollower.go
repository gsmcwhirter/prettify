package streamer

import (
	"context"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gsmcwhirter/prettify/pkg/files/watcher"
	"github.com/gsmcwhirter/prettify/pkg/streams/linehandler"
)

// TailFollower is a stuct that will spawn a goroutine to scan for not-yet-seen files and otherwise behaves like tail -f
type TailFollower struct {
	FileWatcher *watcher.Watcher
	LineHandler linehandler.LineHandler
}

func (tf *TailFollower) watchForNewFiles(ctx context.Context, newFiles chan string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := tf.FileWatcher.Run(ctx)
			if err != nil {
				return err
			}

			for _, path := range tf.FileWatcher.SeenThisTime() {
				newFiles <- path
			}

			time.Sleep(10 * time.Millisecond)
		}
	}
}

// FollowTail kicks off the file watching goroutine and otherwise behaves like tail -f
func (tf *TailFollower) FollowTail(ctx context.Context, lastFile string, lastFilePos int64) (err error) {
	newFiles := make(chan string, 16)
	defer close(newFiles)

	ctx, cancel := context.WithCancel(ctx)

	var g errgroup.Group
	g.Go(func() error {
		err := tf.watchForNewFiles(ctx, newFiles)
		select {
		case <-ctx.Done():
			if err == nil {
				err = ctx.Err()
			}
		default:
			cancel()
		}
		return err
	})

SCAN_LOOP:
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break SCAN_LOOP
		case filePath := <-newFiles:
			dirName, fileName := filepath.Split(filePath)

			if filePath == lastFile {
				lastFilePos, err = CatFrom(ctx, dirName, fileName, lastFilePos, tf.LineHandler)
				if err != nil {
					break
				}
			} else {
				lastFile = filePath
				lastFilePos, err = Cat(ctx, dirName, fileName, tf.LineHandler)
				if err != nil {
					break
				}
			}
		}
	}

	return g.Wait()
}
