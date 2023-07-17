package clouddrive

import (
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/lib/encoder"
	"github.com/rclone/rclone/lib/env"
)

var (
	scopes = []fs.Option{{
		Name: "scope",
		Help: "Scope that rclone should use when requesting access from drive.",
		Examples: []fs.OptionExample{{
			Value: "drive",
			Help:  "Full access all files, excluding Application Data Folder.",
		}, {
			Value: "drive.readonly",
			Help:  "Read-only access to file metadata and file contents.",
		}, {
			Value: "drive.file",
			Help:  "Access to files created by rclone only.\nThese are visible in the drive website.\nFile authorization is revoked when the user deauthorizes the app.",
		}, {
			Value: "drive.appfolder",
			Help:  "Allows read and write access to the Application Data folder.\nThis is not visible in the drive website.",
		}, {
			Value: "drive.metadata.readonly",
			Help:  "Allows read-only access to file metadata but\ndoes not allow any access to read or download file content.",
		}},
	}, {
		Name: "root_folder_id",
		Help: `ID of the root folder.
Leave blank normally.

Fill in to access "Computers" folders (see docs), or for rclone to use
a non root folder as its starting point.
`,
		Advanced: true,
	}, {
		Name: "master_key_file",
		Help: "Path to master key file (JSON)" + env.ShellExpandHelp,
	}, {
		Name:     "auth_owner_only",
		Default:  false,
		Help:     "Only consider files owned by the authenticated user.",
		Advanced: true,
	}, {
		Name:     "use_trash",
		Default:  true,
		Help:     "Send files to the trash instead of deleting permanently.\n\nDefaults to true, namely sending files to the trash.\nUse `--drive-use-trash=false` to delete files permanently instead.",
		Advanced: true,
	}, {
		Name:    "copy_shortcut_content",
		Default: false,
		Help: `Server side copy contents of shortcuts instead of the shortcut.

When doing server side copies, normally rclone will copy shortcuts as
shortcuts.

If this flag is used then rclone will copy the contents of shortcuts
rather than shortcuts themselves when doing server side copies.`,
		Advanced: true,
	}, {
		Name:     "skip_gdocs",
		Default:  false,
		Help:     "Skip google documents in all listings.\n\nIf given, gdocs practically become invisible to rclone.",
		Advanced: true,
	}, {
		Name:    "skip_checksum_gphotos",
		Default: false,
		Help: `Skip MD5 checksum on Google photos and videos only.

Use this if you get checksum errors when transferring Google photos or
videos.

Setting this flag will cause Google photos and videos to return a
blank MD5 checksum.

Google photos are identified by being in the "photos" space.

Corrupted checksums are caused by Google modifying the image/video but
not updating the checksum.`,
		Advanced: true,
	}, {
		Name:    "shared_with_me",
		Default: false,
		Help: `Only show files that are shared with me.

Instructs rclone to operate on your "Shared with me" folder (where
Google Drive lets you access the files and folders others have shared
with you).

This works both with the "list" (lsd, lsl, etc.) and the "copy"
commands (copy, sync, etc.), and with all other commands too.`,
		Advanced: true,
	}, {
		Name:     "trashed_only",
		Default:  false,
		Help:     "Only show files that are in the trash.\n\nThis will show trashed files in their original directory structure.",
		Advanced: true,
	}, {
		Name:     "starred_only",
		Default:  false,
		Help:     "Only show files that are starred.",
		Advanced: true,
	}, {
		Name:     "formats",
		Default:  "",
		Help:     "Deprecated: See export_formats.",
		Advanced: true,
		Hide:     fs.OptionHideConfigurator,
	}, {
		Name:     "export_formats",
		Default:  defaultExportExtensions,
		Help:     "Comma separated list of preferred formats for downloading Google docs.",
		Advanced: true,
	}, {
		Name:     "import_formats",
		Default:  "",
		Help:     "Comma separated list of preferred formats for uploading Google docs.",
		Advanced: true,
	}, {
		Name:     "allow_import_name_change",
		Default:  false,
		Help:     "Allow the filetype to change when uploading Google docs.\n\nE.g. file.doc to file.docx. This will confuse sync and reupload every time.",
		Advanced: true,
	}, {
		Name:    "use_created_date",
		Default: false,
		Help: `Use file created date instead of modified date.

Useful when downloading data and you want the creation date used in
place of the last modified date.

**WARNING**: This flag may have some unexpected consequences.

When uploading to your drive all files will be overwritten unless they
haven't been modified since their creation. And the inverse will occur
while downloading.  This side effect can be avoided by using the
"--checksum" flag.

This feature was implemented to retain photos capture date as recorded
by google photos. You will first need to check the "Create a Google
Photos folder" option in your google drive settings. You can then copy
or move the photos locally and use the date the image was taken
(created) set as the modification date.`,
		Advanced: true,
		Hide:     fs.OptionHideConfigurator,
	}, {
		Name:    "use_shared_date",
		Default: false,
		Help: `Use date file was shared instead of modified date.

Note that, as with "--drive-use-created-date", this flag may have
unexpected consequences when uploading/downloading files.

If both this flag and "--drive-use-created-date" are set, the created
date is used.`,
		Advanced: true,
		Hide:     fs.OptionHideConfigurator,
	}, {
		Name:     "list_chunk",
		Default:  1000,
		Help:     "Size of listing chunk 100-1000, 0 to disable.",
		Advanced: true,
	}, {
		Name:     "impersonate",
		Default:  "",
		Help:     `Impersonate this user when using a service account.`,
		Advanced: true,
	}, {
		Name:    "alternate_export",
		Default: false,
		Help:    "Deprecated: No longer needed.",
		Hide:    fs.OptionHideBoth,
	}, {
		Name:     "upload_cutoff",
		Default:  defaultChunkSize,
		Help:     "Cutoff for switching to chunked upload.",
		Advanced: true,
	}, {
		Name:    "chunk_size",
		Default: defaultChunkSize,
		Help: `Upload chunk size.

Must a power of 2 >= 256k.

Making this larger will improve performance, but note that each chunk
is buffered in memory one per transfer.

Reducing this will reduce memory usage but decrease performance.`,
		Advanced: true,
	}, {
		Name:    "acknowledge_abuse",
		Default: false,
		Help: `Set to allow files which return cannotDownloadAbusiveFile to be downloaded.

If downloading a file returns the error "This file has been identified
as malware or spam and cannot be downloaded" with the error code
"cannotDownloadAbusiveFile" then supply this flag to rclone to
indicate you acknowledge the risks of downloading the file and rclone
will download it anyway.`,
		Advanced: true,
	}, {
		Name:     "keep_revision_forever",
		Default:  false,
		Help:     "Keep new head revision of each file forever.",
		Advanced: true,
	}, {
		Name:    "size_as_quota",
		Default: false,
		Help: `Show sizes as storage quota usage, not actual size.

Show the size of a file as the storage quota used. This is the
current version plus any older versions that have been set to keep
forever.

**WARNING**: This flag may have some unexpected consequences.

It is not recommended to set this flag in your config - the
recommended usage is using the flag form --drive-size-as-quota when
doing rclone ls/lsl/lsf/lsjson/etc only.

If you do use this flag for syncing (not recommended) then you will
need to use --ignore size also.`,
		Advanced: true,
		Hide:     fs.OptionHideConfigurator,
	}, {
		Name:     "v2_download_min_size",
		Default:  fs.SizeSuffix(-1),
		Help:     "If Object's are greater, use drive v2 API to download.",
		Advanced: true,
	}, {
		Name:     "pacer_min_sleep",
		Default:  defaultMinSleep,
		Help:     "Minimum time to sleep between API calls.",
		Advanced: true,
	}, {
		Name:     "pacer_burst",
		Default:  defaultBurst,
		Help:     "Number of API calls to allow without sleeping.",
		Advanced: true,
	}, {
		Name:    "server_side_across_configs",
		Default: false,
		Help: `Allow server-side operations (e.g. copy) to work across different drive configs.

This can be useful if you wish to do a server-side copy between two
different Google drives.  Note that this isn't enabled by default
because it isn't easy to tell if it will work between any two
configurations.`,
		Advanced: true,
	}, {
		Name:    "disable_http2",
		Default: true,
		Help: `Disable drive using http2.

There is currently an unsolved issue with the google drive backend and
HTTP/2.  HTTP/2 is therefore disabled by default for the drive backend
but can be re-enabled here.  When the issue is solved this flag will
be removed.

See: https://github.com/rclone/rclone/issues/3631

`,
		Advanced: true,
	}, {
		Name:    "stop_on_upload_limit",
		Default: false,
		Help: `Make upload limit errors be fatal.

At the time of writing it is only possible to upload 750 GiB of data to
Google Drive a day (this is an undocumented limit). When this limit is
reached Google Drive produces a slightly different error message. When
this flag is set it causes these errors to be fatal.  These will stop
the in-progress sync.

Note that this detection is relying on error message strings which
Google don't document so it may break in the future.

See: https://github.com/rclone/rclone/issues/3857
`,
		Advanced: true,
	}, {
		Name:    "stop_on_download_limit",
		Default: false,
		Help: `Make download limit errors be fatal.

At the time of writing it is only possible to download 10 TiB of data from
Google Drive a day (this is an undocumented limit). When this limit is
reached Google Drive produces a slightly different error message. When
this flag is set it causes these errors to be fatal.  These will stop
the in-progress sync.

Note that this detection is relying on error message strings which
Google don't document so it may break in the future.
`,
		Advanced: true,
	}, {
		Name: "skip_shortcuts",
		Help: `If set skip shortcut files.

Normally rclone dereferences shortcut files making them appear as if
they are the original file (see [the shortcuts section](#shortcuts)).
If this flag is set then rclone will ignore shortcut files completely.
`,
		Advanced: true,
		Default:  false,
	}, {
		Name: "skip_dangling_shortcuts",
		Help: `If set skip dangling shortcut files.

If this is set then rclone will not show any dangling shortcuts in listings.
`,
		Advanced: true,
		Default:  false,
	}, {
		Name: "resource_key",
		Help: `Resource key for accessing a link-shared file.

If you need to access files shared with a link like this

    https://drive.google.com/drive/folders/XXX?resourcekey=YYY&usp=sharing

Then you will need to use the first part "XXX" as the "root_folder_id"
and the second part "YYY" as the "resource_key" otherwise you will get
404 not found errors when trying to access the directory.

See: https://developers.google.com/drive/api/guides/resource-keys

This resource key requirement only applies to a subset of old files.

Note also that opening the folder once in the web interface (with the
user you've authenticated rclone with) seems to be enough so that the
resource key is no needed.
`,
		Advanced: true,
	}, {
		Name:     config.ConfigEncoding,
		Help:     config.ConfigEncodingHelp,
		Advanced: true,
		// Encode invalid UTF-8 bytes as json doesn't handle them properly.
		// Don't encode / as it's a valid name character in drive.
		Default: encoder.EncodeInvalidUtf8,
	}}
)
