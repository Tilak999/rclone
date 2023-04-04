package clouddrive

import "github.com/rclone/rclone/fs"

var CommandHelp = []fs.CommandHelp{{
	Name:  "get",
	Short: "Get command for fetching the drive config parameters",
	Long: `This is a get command which will be used to fetch the various drive config parameters

Usage Examples:

    rclone backend get drive: [-o master_key_file] [-o chunk_size]
    rclone rc backend/command command=get fs=drive: [-o master_key_file] [-o chunk_size]
`,
	Opts: map[string]string{
		"chunk_size":      "show the current upload chunk size",
		"master_key_file": "show the current service account file",
	},
}, {
	Name:  "set",
	Short: "Set command for updating the drive config parameters",
	Long: `This is a set command which will be used to update the various drive config parameters

Usage Examples:

    rclone backend set drive: [-o master_key_file=sa.json] [-o chunk_size=67108864]
    rclone rc backend/command command=set fs=drive: [-o master_key_file=sa.json] [-o chunk_size=67108864]
`,
	Opts: map[string]string{
		"chunk_size":      "update the current upload chunk size",
		"master_key_file": "update the current service account file",
	},
}, {
	Name:  "shortcut",
	Short: "Create shortcuts from files or directories",
	Long: `This command creates shortcuts from files or directories.

Usage:

    rclone backend shortcut drive: source_item destination_shortcut
    rclone backend shortcut drive: source_item -o target=drive2: destination_shortcut

In the first example this creates a shortcut from the "source_item"
which can be a file or a directory to the "destination_shortcut". The
"source_item" and the "destination_shortcut" should be relative paths
from "drive:"

In the second example this creates a shortcut from the "source_item"
relative to "drive:" to the "destination_shortcut" relative to
"drive2:". This may fail with a permission error if the user
authenticated with "drive2:" can't read files from "drive:".
`,
	Opts: map[string]string{
		"target": "optional target remote for the shortcut destination",
	},
}, {
	Name:  "drives",
	Short: "List the Shared Drives available to this account",
	Long: `This command lists the Shared Drives (Team Drives) available to this
account.

Usage:

    rclone backend [-o config] drives drive:

This will return a JSON list of objects like this

    [
        {
            "id": "0ABCDEF-01234567890",
            "kind": "drive#teamDrive",
            "name": "My Drive"
        },
        {
            "id": "0ABCDEFabcdefghijkl",
            "kind": "drive#teamDrive",
            "name": "Test Drive"
        }
    ]

With the -o config parameter it will output the list in a format
suitable for adding to a config file to make aliases for all the
drives found and a combined drive.

    [My Drive]
    type = alias
    remote = drive,team_drive=0ABCDEF-01234567890,root_folder_id=:

    [Test Drive]
    type = alias
    remote = drive,team_drive=0ABCDEFabcdefghijkl,root_folder_id=:

    [AllDrives]
    type = combine
    upstreams = "My Drive=My Drive:" "Test Drive=Test Drive:"

Adding this to the rclone config file will cause those team drives to
be accessible with the aliases shown. Any illegal characters will be
substituted with "_" and duplicate names will have numbers suffixed.
It will also add a remote called AllDrives which shows all the shared
drives combined into one directory tree.
`,
}, {
	Name:  "untrash",
	Short: "Untrash files and directories",
	Long: `This command untrashes all the files and directories in the directory
passed in recursively.

Usage:

This takes an optional directory to trash which make this easier to
use via the API.

    rclone backend untrash drive:directory
    rclone backend -i untrash drive:directory subdir

Use the -i flag to see what would be restored before restoring it.

Result:

    {
        "Untrashed": 17,
        "Errors": 0
    }
`,
}, {
	Name:  "copyid",
	Short: "Copy files by ID",
	Long: `This command copies files by ID

Usage:

    rclone backend copyid drive: ID path
    rclone backend copyid drive: ID1 path1 ID2 path2

It copies the drive file with ID given to the path (an rclone path which
will be passed internally to rclone copyto). The ID and path pairs can be
repeated.

The path should end with a / to indicate copy the file as named to
this directory. If it doesn't end with a / then the last path
component will be used as the file name.

If the destination is a drive backend then server-side copying will be
attempted if possible.

Use the -i flag to see what would be copied before copying.
`,
}, {
	Name:  "exportformats",
	Short: "Dump the export formats for debug purposes",
}, {
	Name:  "importformats",
	Short: "Dump the import formats for debug purposes",
}}
