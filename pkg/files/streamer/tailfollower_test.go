package streamer

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gsmcwhirter/go-util/v9/deferutil"

	"github.com/gsmcwhirter/prettify/pkg/files/pattern"
	"github.com/gsmcwhirter/prettify/pkg/files/watcher"
	"github.com/gsmcwhirter/prettify/pkg/streams/linehandler"
	"github.com/gsmcwhirter/prettify/pkg/testutil"
)

func TestTailFollower_watchForNewFiles(t *testing.T) {
	type fields struct {
		FileWatcher *watcher.Watcher
		LineHandler linehandler.LineHandler
	}
	type args struct {
		newFiles chan string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tf := TailFollower{
				FileWatcher: tt.fields.FileWatcher,
				LineHandler: tt.fields.LineHandler,
			}

			ctx := context.Background()

			err := tf.watchForNewFiles(ctx, tt.args.newFiles)
			if (err == nil) != tt.wantErr {
				t.Errorf("watchForNewFiles error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestTailFollower_FollowTail(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	buffer := testutil.NewPrintfBuffer(1024) // 1Kb to start
	maxDur := 2 * time.Second

	type lpArgs struct {
		withBlanks   bool
		withFilename bool
		withPretty   bool
		withColor    bool
		withSort     bool
		withPath     string
	}
	type args struct {
		lastFile    string
		lastFilePos int64
		maxDuration time.Duration
	}
	tests := []struct {
		name      string
		lpArgs    lpArgs
		args      args
		numFiles  int
		numLines  int
		linebytes []byte
		wantBytes []byte
		wantErr   bool
	}{
		{
			name: "clean lines",
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: true,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				lastFile:    "",
				lastFilePos: 0,
				maxDuration: maxDur,
			},
			numFiles:  3,
			numLines:  5,
			linebytes: []byte("test1test2\ntest3\ntest4\n"),
			wantBytes: []byte(`test-out-0.log: test1test2
test-out-0.log: test3
test-out-0.log: test4
test-out-0.log: test1test2
test-out-0.log: test3
test-out-0.log: test4
test-out-0.log: test1test2
test-out-0.log: test3
test-out-0.log: test4
test-out-0.log: test1test2
test-out-0.log: test3
test-out-0.log: test4
test-out-0.log: test1test2
test-out-0.log: test3
test-out-0.log: test4
test-out-1.log: test1test2
test-out-1.log: test3
test-out-1.log: test4
test-out-1.log: test1test2
test-out-1.log: test3
test-out-1.log: test4
test-out-1.log: test1test2
test-out-1.log: test3
test-out-1.log: test4
test-out-1.log: test1test2
test-out-1.log: test3
test-out-1.log: test4
test-out-1.log: test1test2
test-out-1.log: test3
test-out-1.log: test4
test-out-2.log: test1test2
test-out-2.log: test3
test-out-2.log: test4
test-out-2.log: test1test2
test-out-2.log: test3
test-out-2.log: test4
test-out-2.log: test1test2
test-out-2.log: test3
test-out-2.log: test4
test-out-2.log: test1test2
test-out-2.log: test3
test-out-2.log: test4
test-out-2.log: test1test2
test-out-2.log: test3
test-out-2.log: test4
`),
			wantErr: false,
		},
		{
			name: "non-clean lines",
			lpArgs: lpArgs{
				withBlanks:   false,
				withFilename: false,
				withPretty:   false,
				withColor:    false,
				withSort:     false,
				withPath:     "",
			},
			args: args{
				lastFile:    "",
				lastFilePos: 0,
				maxDuration: maxDur,
			},
			numFiles:  3,
			numLines:  5,
			linebytes: []byte("test1test2\ntest3\ntest4"),
			wantBytes: []byte(`test1test2
test3
test4test1test2
test3
test4test1test2
test3
test4test1test2
test3
test4test1test2
test3
test4
test1test2
test3
test4test1test2
test3
test4test1test2
test3
test4test1test2
test3
test4test1test2
test3
test4
test1test2
test3
test4test1test2
test3
test4test1test2
test3
test4test1test2
test3
test4test1test2
test3
test4
`),
			wantErr: false,
		},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			testDir, err := ioutil.TempDir(os.TempDir(), fmt.Sprintf("FollowTail_Test_%d", i))
			if err != nil {
				t.Error("Could not create temporary test directory")
				return
			}
			defer func() {
				err := os.RemoveAll(testDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not remove directory %s: %s", testDir, err)
				}
			}()

			fp := pattern.NewPattern(fmt.Sprintf("%s/test.log", testDir))

			buffer.Reset()
			lp := linehandler.NewLinePrinter(linehandler.Options{
				WithBlanks:   tt.lpArgs.withBlanks,
				WithFilename: tt.lpArgs.withFilename,
				JSONPath:     tt.lpArgs.withPath,
				Pretty:       tt.lpArgs.withPretty,
				Color:        tt.lpArgs.withColor,
				Sort:         tt.lpArgs.withSort,
				Printf:       buffer.Printf,
			})

			var wg sync.WaitGroup

			writeFile := func(i int) {
				fh, err := os.Create(fmt.Sprintf("%s/test-out-%d.log", testDir, i))
				if err != nil {
					return
				}
				defer deferutil.CheckDefer(fh.Close)

				for j := 0; j < tt.numLines; j++ {
					_, err = fh.Write(tt.linebytes)
					if err != nil {
						return
					}

					time.Sleep(time.Duration(2 * rand.Float64() * float64(time.Millisecond))) //nolint:gosec // don't care about security here
				}

				if tt.linebytes[len(tt.linebytes)-1] != 10 {
					_, err = fh.Write([]byte{10})
					if err != nil {
						return
					}
				}
			}

			writer := func() {
				defer wg.Done()
				for i := 0; i < tt.numFiles; i++ {
					writeFile(i)
				}
			}

			reader := func() {
				done := make(chan struct{})
				defer wg.Done()
				defer func() { // hacky way to
					select {
					case <-done:
					default:
						close(done)
					}
				}()

				tf := TailFollower{
					FileWatcher: watcher.NewWatcher(&fp),
					LineHandler: lp,
				}

				ctx := context.Background()
				if tt.args.maxDuration != 0 {
					ctxt, cancel := context.WithTimeout(ctx, tt.args.maxDuration)
					defer cancel()
					ctx = ctxt
				}

				if err := tf.FollowTail(ctx, tt.args.lastFile, tt.args.lastFilePos); (!errors.Is(err, context.DeadlineExceeded)) && (err != nil) != tt.wantErr {
					t.Errorf("TailFollower.FollowTail() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			wg.Add(1)
			go reader()

			time.Sleep(100 * time.Millisecond)

			wg.Add(1)
			go writer()
			wg.Wait()

			bufferBytes := buffer.GetData()
			if !reflect.DeepEqual(bufferBytes, tt.wantBytes) && (len(bufferBytes) > 0 || len(tt.wantBytes) > 0) {
				t.Errorf("FollowTail() output = %v (\n%s), want %v (\n%s)", bufferBytes, string(bufferBytes), tt.wantBytes, string(tt.wantBytes))
			}
		})
	}
}
