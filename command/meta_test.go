package command

import (
	"flag"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/creack/pty"
	"github.com/hashicorp/nomad/ci"
	cui "github.com/hashicorp/nomad/command/ui"
	"github.com/mattn/go-colorable"
	"github.com/mitchellh/cli"
	"github.com/shoenig/test/must"
	"github.com/stretchr/testify/require"
)

func TestMeta_FlagSet(t *testing.T) {
	ci.Parallel(t)
	cases := []struct {
		Flags    FlagSetFlags
		Expected []string
	}{
		{
			FlagSetNone,
			[]string{},
		},
		{
			FlagSetClient,
			[]string{
				"address",
				"no-color",
				"force-color",
				"region",
				"namespace",
				"ca-cert",
				"ca-path",
				"client-cert",
				"client-key",
				"insecure",
				"tls-server-name",
				"tls-skip-verify",
				"token",
			},
		},
	}

	for i, tc := range cases {
		var m Meta
		fs := m.FlagSet("foo", tc.Flags)

		actual := make([]string, 0, 0)
		fs.VisitAll(func(f *flag.Flag) {
			actual = append(actual, f.Name)
		})
		sort.Strings(actual)
		sort.Strings(tc.Expected)

		if !reflect.DeepEqual(actual, tc.Expected) {
			t.Fatalf("%d: flags: %#v\n\nExpected: %#v\nGot: %#v",
				i, tc.Flags, tc.Expected, actual)
		}
	}
}

func TestMeta_Colorize(t *testing.T) {

	type testCaseSetupFn func(*testing.T, *Meta)

	cases := []struct {
		Name        string
		SetupFn     testCaseSetupFn
		ExpectColor bool
	}{
		{
			Name: "colors if UI is colored",
			SetupFn: func(t *testing.T, m *Meta) {
				m.SetupUi([]string{})
			},
			ExpectColor: true,
		},
		{
			Name: "disable colors via CLI flag",
			SetupFn: func(t *testing.T, m *Meta) {
				m.SetupUi([]string{"-no-color"})
			},
			ExpectColor: false,
		},
		{
			Name: "disable colors via env var",
			SetupFn: func(t *testing.T, m *Meta) {
				t.Setenv(EnvNomadCLINoColor, "1")
				m.SetupUi([]string{})
			},
			ExpectColor: false,
		},
		{
			Name: "force colors via CLI flag",
			SetupFn: func(t *testing.T, m *Meta) {
				m.SetupUi([]string{"-force-color"})
			},
			ExpectColor: true,
		},
		{
			Name: "force colors via env var",
			SetupFn: func(t *testing.T, m *Meta) {
				t.Setenv(EnvNomadCLIForceColor, "1")
				m.SetupUi([]string{})
			},
			ExpectColor: true,
		},
		{
			Name: "no color take predecence over force color via CLI flag",
			SetupFn: func(t *testing.T, m *Meta) {
				m.SetupUi([]string{"-no-color", "-force-color"})
			},
			ExpectColor: false,
		},
		{
			Name: "no color take predecence over force color via env var",
			SetupFn: func(t *testing.T, m *Meta) {
				t.Setenv(EnvNomadCLINoColor, "1")
				m.SetupUi([]string{"-force-color"})
			},
			ExpectColor: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Create fake test terminal.
			_, tty, err := pty.Open()
			require.NoError(t, err)
			defer tty.Close()

			oldStdout := os.Stdout
			defer func() { os.Stdout = oldStdout }()
			os.Stdout = tty

			// Make sure color related environment variables are clean.
			t.Setenv(EnvNomadCLIForceColor, "")
			t.Setenv(EnvNomadCLINoColor, "")

			// Run test case.
			m := &Meta{}
			if tc.SetupFn != nil {
				tc.SetupFn(t, m)
			}
			cui, _ := m.Ui.(*cui.ColoredUi)
			bui, _ := cui.Ui.(*cli.BasicUi)

			if !tc.ExpectColor {
				require.IsType(t, bui.Writer, &colorable.NonColorable{})
				require.IsType(t, bui.Writer, &colorable.NonColorable{})
			} else {
				require.IsType(t, bui.Writer, &os.File{})
				require.IsType(t, bui.Writer, &os.File{})
			}
		})
	}
}

func TestMeta_ColorizeStderrWhenStdoutNotATTY(t *testing.T) {
	// Create fake test terminal.
	_, tty, err := pty.Open()

	must.NoError(t, err)
	defer tty.Close()

	oldStderr := os.Stderr
	oldStdout := os.Stdout
	oFile, err := os.CreateTemp(t.TempDir(), "test")
	must.NoError(t, err)

	defer func() {
		os.Stderr = oldStderr
		os.Stdout = oldStdout
	}()
	os.Stdout = oFile
	os.Stderr = tty

	// Make sure color related environment variables are clean.
	t.Setenv(EnvNomadCLIForceColor, "")
	t.Setenv(EnvNomadCLINoColor, "")

	// Run test case.
	m := &Meta{}
	m.SetupUi([]string{})

	cui, _ := m.Ui.(*cui.ColoredUi)
	bui, _ := cui.Ui.(*cli.BasicUi)

	require.IsType(t, bui.Writer, &colorable.NonColorable{})
	require.IsType(t, bui.ErrorWriter, &os.File{})
}

func TestMeta_ColorizeStdoutWhenStderrNotATTY(t *testing.T) {
	// Create fake test terminal.
	_, tty, err := pty.Open()

	must.NoError(t, err)
	defer tty.Close()

	oldStderr := os.Stderr
	oldStdout := os.Stdout
	oFile, err := os.CreateTemp(t.TempDir(), "test")
	must.NoError(t, err)

	defer func() {
		os.Stderr = oldStderr
		os.Stdout = oldStdout
	}()
	os.Stdout = tty
	os.Stderr = oFile

	// Make sure color related environment variables are clean.
	t.Setenv(EnvNomadCLIForceColor, "")
	t.Setenv(EnvNomadCLINoColor, "")

	// Run test case.
	m := &Meta{}
	m.SetupUi([]string{})

	cui, _ := m.Ui.(*cui.ColoredUi)
	bui, _ := cui.Ui.(*cli.BasicUi)

	require.IsType(t, bui.Writer, &os.File{})
	require.IsType(t, bui.ErrorWriter, &colorable.NonColorable{})
}

func TestMeta_ColorizeNoneWhenNotATTY(t *testing.T) {
	oldStderr := os.Stderr
	oldStdout := os.Stdout
	td := t.TempDir()
	oFile, err := os.CreateTemp(td, "test")
	must.NoError(t, err)

	defer func() {
		os.Stderr = oldStderr
		os.Stdout = oldStdout
	}()
	os.Stdout = oFile
	os.Stderr = oFile

	// Make sure color related environment variables are clean.
	t.Setenv(EnvNomadCLIForceColor, "")
	t.Setenv(EnvNomadCLINoColor, "")

	// Run test case.
	m := &Meta{}
	m.SetupUi([]string{})

	cui, _ := m.Ui.(*cui.ColoredUi)
	bui, _ := cui.Ui.(*cli.BasicUi)

	require.IsType(t, bui.Writer, &colorable.NonColorable{})
	require.IsType(t, bui.ErrorWriter, &colorable.NonColorable{})
}
