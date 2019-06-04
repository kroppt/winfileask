package winfileask

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

var (
	modcomdlg32         = syscall.NewLazyDLL("comdlg32.dll")
	procGetSaveFileName = modcomdlg32.NewProc("GetSaveFileNameW")
	procGetOpenFileName = modcomdlg32.NewProc("GetOpenFileNameW")
)

// The flags for the Flags member of TagOFNA.
const (
	// AllowMultiSelect means the File Name list box allows multiple
	// selections. If you also set the Explorer flag, the dialog box uses the
	// Explorer-style user interface; otherwise, it uses the old-style user
	// interface.
	//
	// If the user selects more than one file, the lpstrFile buffer returns the
	// path to the current directory followed by the file names of the selected
	// files. The nFileOffset member is the offset, in bytes or characters, to
	// the first file name, and the nFileExtension member is not used. For
	// Explorer-style dialog boxes, the directory and file name strings are
	// NULL separated, with an extra NULL character after the last file name.
	// This format enables the Explorer-style dialog boxes to return long file
	// names that include spaces. For old-style dialog boxes, the directory and
	// file name strings are separated by spaces and the function uses short
	// file names for file names with spaces. You can use the FindFirstFile
	// (https://msdn.microsoft.com/02fc92c4-582d-4c9f-a811-b5c839e9fffa)
	// function to convert between long and short file names.
	//
	// If you specify a custom template for an old-style dialog box, the
	// definition of the File Name list box must contain the LBS_EXTENDEDSEL
	// value.
	AllowMultiSelect uint32 = 0x00000200
	// CreatePrompt causes the dialog box to prompt the user for permission to
	// create the file if the user specifies a file that does not exist. If
	// the user chooses to create the file, the dialog box closes and the
	// function returns the specified name; otherwise, the dialog box remains
	// open. If you use this flag with the AllowMultiSelect flag, the dialog
	// box allows the user to specify only one nonexistent file.
	CreatePrompt uint32 = 0x00002000
	// DontAddToRecent prevents the system from adding a link to the selected
	// file in the file system directory that contains the user's most recently
	// used documents. To retrieve the location of this directory, call the
	// SHGetSpecialFolderLocation
	// (https://msdn.microsoft.com/en-us/library/Bb762203(v=VS.85).aspx)
	// function with the CSIDL_RECENT flag.
	DontAddToRecent uint32 = 0x02000000
	// EnableHook enables the hook function specified in the lpfnHook member.
	EnableHook uint32 = 0x00000020
	// EnableIncludeNotify causes the dialog box to send CDN_INCLUDEITEM
	// (https://msdn.microsoft.com/en-us/library/ms646862(v=VS.85).aspx)
	// notification messages to your OFNHookProc
	// (https://msdn.microsoft.com/en-us/library/ms646931(v=VS.85).aspx)
	// hook procedure when the user opens a folder. The dialog box sends a
	// notification for each item in the newly opened folder. These messages
	// enable you to control which items the dialog box displays in the
	// folder's item list.
	EnableIncludeNotify uint32 = 0x00400000
	// EnableSizing enables the Explorer-style dialog box to be resized using
	// either the mouse or the keyboard. By default, the Explorer-style Open
	// and Save As dialog boxes allow the dialog box to be resized regardless
	// of whether this flag is set. This flag is necessary only if you provide
	// a hook procedure or custom template. The old-style dialog box does not
	// permit resizing.
	EnableSizing uint32 = 0x00800000
	// EnableTemplate means the lpTemplateName member is a pointer to the name
	// of a dialog template resource in the module identified by the hInstance
	// member. If the Explorer flag is set, the system uses the specified
	// template to create a dialog box that is a child of the default Explorer-
	// style dialog box. If the Explorer flag is not set, the system uses the
	// template to create an old-style dialog box that replaces the default
	// dialog box.
	EnableTemplate uint32 = 0x00000040
	// EnableTemplateHandle means the hInstance member identifies a data block
	// that contains a preloaded dialog box template. The system ignores
	// lpTemplateName if this flag is specified. If the Explorer flag is set,
	// the system uses the specified template to create a dialog box that is a
	// child of the default Explorer-style dialog box. If the Explorer flag is
	// not set, the system uses the template to create an old-style dialog box
	// that replaces the default dialog box.
	EnableTemplateHandle uint32 = 0x00000080
	// Explorer indicates that any customizations made to the Open or Save As
	// dialog box use the Explorer-style customization methods. For more
	// information, see Explorer-Style Hook Procedures
	// (https://msdn.microsoft.com/en-us/library/ms646960(v=VS.85).aspx)
	// and Explorer-Style Custom Templates
	// (https://msdn.microsoft.com/en-us/library/ms646960(v=VS.85).aspx).
	// By default, the Open and Save As dialog boxes use the Explorer-style
	// user interface regardless of whether this flag is set. This flag is
	// necessary only if you provide a hook procedure or custom template, or
	// set the AllowMultiSelect flag.
	// If you want the old-style user interface, omit the Explorer flag and
	// provide a replacement old-style template or hook procedure. If you want
	// the old style but do not need a custom template or hook procedure,
	// simply provide a hook procedure that always returns FALSE.
	Explorer uint32 = 0x00080000
	// ExtensionDifferent means the user typed a file name extension that
	// differs from the extension specified by lpstrDefExt. The function does
	// not use this flag if lpstrDefExt is NULL.
	ExtensionDifferent uint32 = 0x00000400
	// FileMustExist means the user can type only names of existing files in
	// the File Name entry field. If this flag is specified and the user enters
	// an invalid name, the dialog box procedure displays a warning in a
	// message box. If this flag is specified, the PathMustExist flag is also
	// used. This flag can be used in an Open dialog box. It cannot be used
	// with a Save As dialog box.
	FileMustExist uint32 = 0x00001000
	// ForceShowHidden forces the showing of system and hidden files, thus
	// overriding the user setting to show or not show hidden files. However,
	// a file that is marked both system and hidden is not shown.
	ForceShowHidden uint32 = 0x10000000
	// HideReadOnly hides the Read Only check box.
	HideReadOnly uint32 = 0x00000004
	// LongNames causes the dialog box to use long file names for old-style
	// dialog boxes. If this flag is not specified, or if the AllowMultiSelect
	// flag is also set, old-style dialog boxes use short file names
	// (8.3 format) for file names with spaces. Explorer-style dialog boxes
	// ignore this flag and always display long file names.
	LongNames uint32 = 0x00200000
	// NoChangeDir restores the current directory to its original value if the
	// user changed the directory while searching for files.
	// This flag is ineffective for GetOpenFileName
	// (https://msdn.microsoft.com/en-us/library/ms646927(v=VS.85).aspx).
	NoChangeDir uint32 = 0x00000008
	// NoDereferenceLinks directs the dialog box to return the path and file
	// name of the selected shortcut (.LNK) file. If this value is not
	// specified, the dialog box returns the path and file name of the file
	// referenced by the shortcut.
	NoDereferenceLinks uint32 = 0x00100000
	// NoLongNames causes the dialog box to use short file names (8.3 format)
	// for old-style dialog boxes. Explorer-style dialog boxes ignore this flag
	// and always display long file names.
	NoLongNames uint32 = 0x00040000
	// NoNetworkButton hides and disables the Network button.
	NoNetworkButton uint32 = 0x00020000
	// NoReadOnlyReturn means the returned file does not have the Read Only
	// check box selected and is not in a write-protected directory.
	NoReadOnlyReturn uint32 = 0x00008000
	// NoTestFileCreate means the file is not created before the dialog box is
	// closed. This flag should be specified if the application saves the file
	// on a create-nonmodify network share. When an application specifies this
	// flag, the library does not check for write protection, a full disk, an
	// open drive door, or network protection. Applications using this flag
	// must perform file operations carefully, because a file cannot be
	// reopened once it is closed.
	NoTestFileCreate uint32 = 0x00010000
	// NoValidate means the common dialog boxes allow invalid characters in the
	// returned file name. Typically, the calling application uses a hook
	// procedure that checks the file name by using the FILEOKSTRING
	// (https://msdn.microsoft.com/en-us/library/ms646870(v=VS.85).aspx)
	// message. If the text box in the edit control is empty or contains
	// nothing but spaces, the lists of files and directories are updated. If
	// the text box in the edit control contains anything else, nFileOffset and
	// nFileExtension are set to values generated by parsing the text. No
	// default extension is added to the text, nor is text copied to the buffer
	// specified by lpstrFileTitle. If the value specified by nFileOffset is
	// less than zero, the file name is invalid. Otherwise, the file name is
	// valid, and nFileExtension and nFileOffset can be used as if the
	// NoValidate flag had not been specified.
	NoValidate uint32 = 0x00000100
	// OverwritePrompt causes the Save As dialog box to generate a message box
	// if the selected file already exists. The user must confirm whether to
	// overwrite the file.
	OverwritePrompt uint32 = 0x00000002
	// PathMustExist means the user can type only valid paths and file names.
	// If this flag is used and the user types an invalid path and file name
	// in the File Name entry field, the dialog box function displays a
	// warning in a message box.
	PathMustExist uint32 = 0x00000800
	// ReadOnly causes the Read Only check box to be selected initially when
	// the dialog box is created. This flag indicates the state of the Read
	// Only check box when the dialog box is closed.
	ReadOnly uint32 = 0x00000001
	// ShareAware specifies that if a call to the OpenFile
	// (https://msdn.microsoft.com/800f4d40-252a-44fe-b10d-348c22d69355)
	// function fails because of a network sharing violation, the error is
	// ignored and the dialog box returns the selected file name. If this flag
	// is not set, the dialog box notifies your hook procedure when a network
	// sharing violation occurs for the file name specified by the user. If
	// you set the Explorer flag, the dialog box sends the CDN_SHAREVIOLATION
	// (https://msdn.microsoft.com/en-us/library/ms646866(v=VS.85).aspx)
	// message to the hook procedure. If you do not set Explorer, the dialog
	// box sends the SHAREVISTRING
	// (https://msdn.microsoft.com/en-us/library/ms646878(v=VS.85).aspx)
	// registered message to the hook procedure.
	ShareAware uint32 = 0x00004000
	// ShowHelp causes the dialog box to display the Help button. The hwndOwner
	// member must specify the window to receive the HELPMSGSTRING
	// (https://msdn.microsoft.com/en-us/library/ms646874(v=VS.85).aspx)
	// registered messages that the dialog box sends when the user clicks the
	// Help button. An Explorer-style dialog box sends a CDN_HELP
	// (https://msdn.microsoft.com/en-us/library/ms646860(v=VS.85).aspx)
	// notification message to your hook procedure when the user clicks the
	// Help button.
	ShowHelp uint32 = 0x00000010
)

// Extra flag definitions.
const (
	// ExNoPlacesBar means the places bar is not displayed. If this flag is not
	// set, Explorer-style dialog boxes include a places bar containing icons
	// for commonly-used folders, such as Favorites and Desktop.
	ExNoPlacesBar = 0x00000001
)

// TagOFNA contains information that the GetOpenFileName
// (https://msdn.microsoft.com/en-us/library/ms646927(v=VS.85).aspx)
// and GetSaveFileName
// (https://msdn.microsoft.com/en-us/library/ms646928(v=VS.85).aspx)
// functions use to initialize an Open or Save As dialog box. After the user
// closes the dialog box, the system returns information about the user's
// selection in this structure.
// https://docs.microsoft.com/en-us/windows/desktop/api/commdlg/ns-commdlg-tagofna
// Remarks:
//    For compatibility reasons, the Places Bar is hidden if Flags is set to
//    EnableHook and lStructSize is OPENFILENAME_SIZE_VERSION_400.
// Minimum supported client:
//    Windows 2000 Professional [desktop apps only]
// Minimum supported server:
//    Windows 2000 Server [desktop apps only]
type TagOFNA struct {
	// The length, in bytes, of the structure. Use `sizeof (OPENFILENAME)` for
	// this parameter.
	LStructSize uint32
	// A handle to the window that owns the dialog box. This member can be any
	// valid window handle, or it can be NULL if the dialog box has no owner.
	HwndOwner unsafe.Pointer
	// If the EnableTemplateHandle flag is set in the Flags member, hInstance
	// is a handle to a memory object containing a dialog box template. If the
	// EnableTemplate flag is set, hInstance is a handle to a module that
	// contains a dialog box template named by the lpTemplateName member. If
	// neither flag is set, this member is ignored. If the Explorer flag is
	// set, the system uses the specified template to create a dialog box that
	// is a child of the default Explorer-style dialog box. If the Explorer
	// flag is not set, the system uses the template to create an old-style
	// dialog box that replaces the default dialog box.
	HInstance unsafe.Pointer // not implemented
	// A buffer containing pairs of null-terminated filter strings. The last
	// string in the buffer must be terminated by two NULL characters.
	//
	// The first string in each pair is a display string that describes the
	// filter (for example, "Text Files"), and the second string specifies the
	// filter pattern (for example, ".TXT"). To specify multiple filter
	// patterns for a single display string, use a semicolon to separate the
	// patterns (for example, ".TXT;.DOC;.BAK"). A pattern string can be a
	// combination of valid file name characters and the asterisk (*) wildcard
	// character. Do not include spaces in the pattern string.
	//
	// The system does not change the order of the filters. It displays them
	// in the File Types combo box in the order specified in lpstrFilter.
	//
	// If lpstrFilter is NULL, the dialog box does not display any filters.
	//
	// In the case of a shortcut, if no filter is set, GetOpenFileName
	// (https://msdn.microsoft.com/en-us/library/ms646927(v=VS.85).aspx)
	// and GetSaveFileName
	// (https://msdn.microsoft.com/en-us/library/ms646928(v=VS.85).aspx)
	// retrieve the name of the .lnk file, not its target. This behavior is the
	// same as setting the NoDereferenceLinks flag in the Flags member. To
	// retrieve a shortcut's target without filtering, use the string
	// `"All Files\0*.*\0\0"`.
	LpstrFilter *uint16
	// A static buffer that contains a pair of null-terminated filter strings
	// for preserving the filter pattern chosen by the user. The first string
	// is your display string that describes the custom filter, and the second
	// string is the filter pattern selected by the user. The first time your
	// application creates the dialog box, you specify the first string, which
	// can be any nonempty string. When the user selects a file, the dialog box
	// copies the current filter pattern to the second string. The preserved
	// filter pattern can be one of the patterns specified in the lpstrFilter
	// buffer, or it can be a filter pattern typed by the user. The system uses
	// the strings to initialize the user-defined file filter the next time the
	// dialog box is created. If the nFilterIndex member is zero, the dialog
	// box uses the custom filter.
	//
	// If this member is NULL, the dialog box does not preserve user-defined
	// filter patterns.
	//
	// If this member is not NULL, the value of the nMaxCustFilter member must
	// specify the size, in characters, of the lpstrCustomFilter buffer.
	LpstrCustomFilter *uint16 // not implemented
	// The size, in characters, of the buffer identified by lpstrCustomFilter.
	// This buffer should be at least 40 characters long. This member is
	// ignored if lpstrCustomFilter is NULL or points to a NULL string.
	NMaxCustFilter uint32 // not implemented
	// The index of the currently selected filter in the File Types control.
	// The buffer pointed to by lpstrFilter contains pairs of strings that
	// define the filters. The first pair of strings has an index value of 1,
	// the second pair 2, and so on. An index of zero indicates the custom
	// filter specified by lpstrCustomFilter. You can specify an index on input
	// to indicate the initial filter description and filter pattern for the
	// dialog box. When the user selects a file, nFilterIndex returns the index
	// of the currently displayed filter. If nFilterIndex is zero and
	// lpstrCustomFilter is NULL, the system uses the first filter in the
	// lpstrFilter buffer. If all three members are zero or NULL, the system
	// does not use any filters and does not show any files in the file list
	// control of the dialog box.
	NFilterIndex uint32
	// The file name used to initialize the File Name edit control. The first
	// character of this buffer must be NULL if initialization is not
	// necessary. When the GetOpenFileName
	// (https://msdn.microsoft.com/en-us/library/ms646927(v=VS.85).aspx)
	// or GetSaveFileName
	// (https://msdn.microsoft.com/en-us/library/ms646928(v=VS.85).aspx)
	// function returns successfully, this buffer contains the drive
	// designator, path, file name, and extension of the selected file.
	//
	// If the AllowMultiSelect flag is set and the user selects multiple files,
	// the buffer contains the current directory followed by the file names of
	// the selected files. For Explorer-style dialog boxes, the directory and
	// file name strings are NULL separated, with an extra NULL character after
	// the last file name. For old-style dialog boxes, the strings are space
	// separated and the function uses short file names for file names with
	// spaces. You can use the FindFirstFile
	// (https://msdn.microsoft.com/02fc92c4-582d-4c9f-a811-b5c839e9fffa)
	// function to convert between long and short file names. If the user
	// selects only one file, the lpstrFile string does not have a separator
	// between the path and file name.
	//
	// If the buffer is too small, the function returns FALSE and the
	// CommDlgExtendedError
	// (https://msdn.microsoft.com/en-us/library/ms646916(v=VS.85).aspx)
	// function returns FNERR_BUFFERTOOSMALL. In this case, the first two bytes
	// of the lpstrFile buffer contain the required size, in bytes or
	// characters.
	LpstrFile *uint16
	// The size, in characters, of the buffer pointed to by lpstrFile. The
	// buffer must be large enough to store the path and file name string or
	// strings, including the terminating NULL character. The GetOpenFileName
	// (https://msdn.microsoft.com/en-us/library/ms646927(v=VS.85).aspx)
	// and GetSaveFileName
	// (https://msdn.microsoft.com/en-us/library/ms646928(v=VS.85).aspx)
	// functions return FALSE if the buffer is too small to contain the file
	// information. The buffer should be at least 256 characters long.
	NMaxFile uint32
	// The file name and extension (without path information) of the selected
	// file. This member can be NULL.
	LpstrFileTitle *uint16 // not implemented
	// The size, in characters, of the buffer pointed to by lpstrFileTitle.
	// This member is ignored if lpstrFileTitle is NULL.
	NMaxFileTitle uint32 // not implemented
	// The initial directory. The algorithm for selecting the initial directory
	// varies on different platforms.
	LpstrInitialDir *uint16
	// A string to be placed in the title bar of the dialog box. If this member
	// is NULL, the system uses the default title (that is, Save As or Open).
	LpstrTitle *uint16
	// A set of bit flags you can use to initialize the dialog box. When the
	// dialog box returns, it sets these flags to indicate the user's input.
	// This member can be a combination of the package flags.
	Flags uint32
	// The zero-based offset, in characters, from the beginning of the path to
	// the file name in the string pointed to by lpstrFile. For the ANSI
	// version, this is the number of bytes; for the Unicode version, this is
	// the number of characters. For example, if lpstrFile points to the
	// following string, "c:\dir1\dir2\file.ext", this member contains the
	// value 13 to indicate the offset of the "file.ext" string. If the user
	// selects more than one file, nFileOffset is the offset to the first file
	// name.
	NFileOffset uint16
	// The zero-based offset, in characters, from the beginning of the path to
	// the file name extension in the string pointed to by lpstrFile. For the
	// ANSI version, this is the number of bytes; for the Unicode version, this
	// is the number of characters. Usually the file name extension is the
	// substring which follows the last occurrence of the dot (".") character.
	// For example, txt is the extension of the filename readme.txt, html the
	// extension of readme.txt.html. Therefore, if lpstrFile points to the
	// string "c:\dir1\dir2\readme.txt", this member contains the value 20. If
	// lpstrFile points to the string "c:\dir1\dir2\readme.txt.html", this
	// member contains the value 24. If lpstrFile points to the string
	// "c:\dir1\dir2\readme.txt.html.", this member contains the value 29. If
	// lpstrFile points to a string that does not contain any "." character
	// such as "c:\dir1\dir2\readme", this member contains zero.
	NFileExtension uint16
	// The default extension. GetOpenFileName
	// (https://msdn.microsoft.com/en-us/library/ms646927(v=VS.85).aspx)
	// and GetSaveFileName
	// (https://msdn.microsoft.com/en-us/library/ms646928(v=VS.85).aspx)
	// append this extension to the file name if the user fails to type an
	// extension. This string can be any length, but only the first three
	// characters are appended. The string should not contain a period (.). If
	// this member is NULL and the user fails to type an extension, no
	// extension is appended.
	LpstrDefExt *uint16 // not implemented
	// Application-defined data that the system passes to the hook procedure
	// identified by the lpfnHook member. When the system sends the
	// WM_INITDIALOG
	// (https://msdn.microsoft.com/en-us/library/ms645428(v=VS.85).aspx)
	// message to the hook procedure, the message's lParam parameter is a
	// pointer to the OPENFILENAME (TagOFNA) structure specified when the
	// dialog box was created. The hook procedure can use this pointer to get
	// the lCustData value.
	LCustData uintptr // not implemented
	// A pointer to a hook procedure. This member is ignored unless the Flags
	// member includes the EnableHook flag.
	//
	// If the Explorer flag is not set in the Flags member, lpfnHook is a
	// pointer to an OFNHookProcOldStyle
	// (https://msdn.microsoft.com/ee551824-51f9-422d-9741-96248e3fc8cc)
	// hook procedure that receives messages intended for the dialog box. The
	// hook procedure returns FALSE to pass a message to the default dialog box
	// procedure or TRUE to discard the message.
	//
	// If Explorer is set, lpfnHook is a pointer to an OFNHookProc
	// (https://msdn.microsoft.com/en-us/library/ms646931(v=VS.85).aspx)
	// hook procedure. The hook procedure receives notification messages sent
	// from the dialog box. The hook procedure also receives messages for any
	// additional controls that you defined by specifying a child dialog
	// template. The hook procedure does not receive messages intended for the
	// standard controls of the default dialog box.
	LpfnHook uintptr // not implemented
	// The name of the dialog template resource in the module identified by the
	// hInstance member. For numbered dialog box resources, this can be a value
	// returned by the MAKEINTRESOURCE
	// (https://msdn.microsoft.com/en-us/library/ms648029(v=VS.85).aspx) macro.
	// This member is ignored unless the EnableTemplate flag is set in the
	// Flags member. If the Explorer flag is set, the system uses the specified
	// template to create a dialog box that is a child of the default Explorer-
	// style dialog box. If the OFN_EXPLORER flag is not set, the system uses
	// the template to create an old-style dialog box that replaces the default
	// dialog box.
	LpTemplateName *uint16 // not implemented
	// This member is reserved.
	PvReserved unsafe.Pointer // not implemented
	// This member is reserved.
	DwReserved uint32 // not implemented
	// A set of bit flags you can use to initialize the dialog box. Currently,
	// this member can be zero or the ExNoPlacesBar flag.
	FlagsEx uint32 // not implemented
}

// Filter represents a file filter and its name and pattern.
type Filter struct {
	Name    string
	Pattern string
}

// FileFilter is a list of Filters.
type FileFilter []Filter

// ToRaw returns a uint16 pointer to the string representation of the filter.
func (ff *FileFilter) ToRaw() (*uint16, error) {
	var sb strings.Builder
	var ptr []uint16
	var err error
	for _, f := range *ff {
		sb.WriteString(f.Name)
		sb.WriteRune('|')
		if strings.ContainsRune(f.Pattern, ' ') {
			return nil, fmt.Errorf("pattern contains a space")
		}
		sb.WriteString(f.Pattern)
		sb.WriteRune('|')
	}
	sb.WriteRune('|')
	if ptr, err = syscall.UTF16FromString(sb.String()); err != nil {
		return nil, err
	}
	for i := range ptr {
		if ptr[i] == uint16('|') {
			ptr[i] = uint16(0)
		}
	}
	return &ptr[0], nil
}

// NewTagOFNA returns an initialized TagOFNA struct
func NewTagOFNA(parentHWND unsafe.Pointer, title string, filter FileFilter, initialDir string, flags uint32) (*TagOFNA, error) {
	var ofn TagOFNA
	var lStructSize uint32
	lStructSize = uint32(unsafe.Sizeof(ofn))
	var lpstrTitle *uint16
	var err error
	if lpstrTitle, err = syscall.UTF16PtrFromString(title); err != nil {
		return nil, err
	}
	var lpstrFilter *uint16
	if lpstrFilter, err = filter.ToRaw(); err != nil {
		return nil, err
	}
	var lpstrInitialDir *uint16
	if lpstrInitialDir, err = syscall.UTF16PtrFromString(initialDir); err != nil {
		return nil, err
	}
	return &TagOFNA{
		LStructSize:     lStructSize,
		HwndOwner:       parentHWND,
		LpstrFilter:     lpstrFilter,
		NFilterIndex:    0,   // defaults to first filter
		LpstrFile:       nil, // set by user
		NMaxFile:        0,   // set by user
		LpstrInitialDir: lpstrInitialDir,
		LpstrTitle:      lpstrTitle,
		Flags:           flags,
		NFileOffset:     0, // set by system
		NFileExtension:  0, // set by system
	}, nil
}

// GetOpenFileName creates an Open dialog box that lets the user specify the
// drive, directory, and the name of a file or set of files to be opened.
func GetOpenFileName(parentHWND unsafe.Pointer, title string, filter FileFilter, initialDir string) (string, bool, error) {
	var ofn *TagOFNA
	var err error
	flags := FileMustExist | HideReadOnly | PathMustExist
	if ofn, err = NewTagOFNA(parentHWND, title, filter, initialDir, flags); err != nil {
		return "", false, err
	}
	buf := make([]uint16, 1024)
	ofn.LpstrFile = &buf[0]
	ofn.NMaxFile = 1024
	ret, _, _ := procGetOpenFileName.Call(uintptr(unsafe.Pointer(ofn)))
	if ret == 0 {
		return "", false, nil
	}
	str := syscall.UTF16ToString(buf)
	return str, true, nil
}

// GetSaveFileName creates a Save dialog box that lets the user specify the
// drive, directory, and name of a file to save.
func GetSaveFileName(parentHWND unsafe.Pointer, title string, filter FileFilter, initialDir string) (string, bool, error) {
	var ofn *TagOFNA
	var err error
	flags := HideReadOnly | PathMustExist
	if ofn, err = NewTagOFNA(parentHWND, title, filter, initialDir, flags); err != nil {
		return "", false, err
	}
	buf := make([]uint16, 1024)
	ofn.LpstrFile = &buf[0]
	ofn.NMaxFile = 1024
	ret, _, _ := procGetSaveFileName.Call(uintptr(unsafe.Pointer(ofn)))
	if ret == 0 {
		return "", false, nil
	}
	str := syscall.UTF16ToString(buf)
	return str, true, nil
}
