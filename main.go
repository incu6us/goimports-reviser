package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/incu6us/goimports-reviser/v3/helper"
	"github.com/incu6us/goimports-reviser/v3/reviser"
)

const (
	projectNameArg         = "project-name"
	versionArg             = "version"
	versionOnlyArg         = "version-only"
	removeUnusedImportsArg = "rm-unused"
	setAliasArg            = "set-alias"
	companyPkgPrefixesArg  = "company-prefixes"
	outputArg              = "output"
	importsOrderArg        = "imports-order"
	formatArg              = "format"
	listDiffFileNameArg    = "list-diff"
	setExitStatusArg       = "set-exit-status"
	recursiveArg           = "recursive"
	useCacheArg            = "use-cache"
	applyToGeneratedFiles  = "apply-to-generated-files"
	excludesArg            = "excludes"
	// using a regex here so that this will work with forked repos (at least on github.com)
	modulePathRegex = `^github.com/[\w-]+/goimports-reviser(/v\d+)?@?`

	// Deprecated options
	localArg    = "local"
	filePathArg = "file-path"
)

// Project build specific vars
var (
	Tag       string
	Commit    string
	SourceURL string
	GoVersion string

	shouldShowVersion           *bool
	shouldShowVersionOnly       *bool
	shouldRemoveUnusedImports   *bool
	shouldSetAlias              *bool
	shouldFormat                *bool
	shouldApplyToGeneratedFiles *bool
	listFileName                *bool
	setExitStatus               *bool
	isRecursive                 *bool
	isUseCache                  *bool
	modulePathMatcher           = regexp.MustCompile(modulePathRegex)
)

var (
	projectName, companyPkgPrefixes, output, importsOrder, excludes string

	// Deprecated
	localPkgPrefixes, filePath string
)

func init() {
	flag.StringVar(
		&filePath,
		filePathArg,
		"",
		"Deprecated. Put file name as an argument(last item) of command line.",
	)

	flag.StringVar(
		&excludes,
		excludesArg,
		"",
		"Exclude files or dirs, example: '.git/,proto/*.go'.",
	)

	flag.StringVar(
		&projectName,
		projectNameArg,
		"",
		"Your project name(ex.: github.com/incu6us/goimports-reviser). Optional parameter.",
	)

	flag.StringVar(
		&companyPkgPrefixes,
		companyPkgPrefixesArg,
		"",
		"Company package prefixes which will be placed after 3rd-party group by default(if defined). Values should be comma-separated. Optional parameters.",
	)

	flag.StringVar(
		&localPkgPrefixes,
		localArg,
		"",
		"Deprecated",
	)

	flag.StringVar(
		&output,
		outputArg,
		"file",
		`Can be "file", "write" or "stdout". Whether to write the formatted content back to the file or to stdout. When "write" together with "-list-diff" will list the file name and write back to the file. Optional parameter.`,
	)

	flag.StringVar(
		&importsOrder,
		importsOrderArg,
		"std,general,company,project",
		`Your imports groups can be sorted in your way. 
std - std import group; 
general - libs for general purpose; 
company - inter-org or your company libs(if you set '-company-prefixes'-option, then 4th group will be split separately. In other case, it will be the part of general purpose libs); 
project - your local project dependencies;
blanked - imports with "_" alias;
dotted - imports with "." alias.
Optional parameter.`,
	)

	listFileName = flag.Bool(
		listDiffFileNameArg,
		false,
		"Option will list files whose formatting differs from goimports-reviser. Optional parameter.",
	)

	setExitStatus = flag.Bool(
		setExitStatusArg,
		false,
		"set the exit status to 1 if a change is needed/made. Optional parameter.",
	)

	shouldRemoveUnusedImports = flag.Bool(
		removeUnusedImportsArg,
		false,
		"Remove unused imports. Optional parameter.",
	)

	shouldSetAlias = flag.Bool(
		setAliasArg,
		false,
		"Set alias for versioned package names, like 'github.com/go-pg/pg/v9'. "+
			"In this case import will be set as 'pg \"github.com/go-pg/pg/v9\"'. Optional parameter.",
	)

	shouldFormat = flag.Bool(
		formatArg,
		false,
		"Option will perform additional formatting. Optional parameter.",
	)

	isRecursive = flag.Bool(
		recursiveArg,
		false,
		"Apply rules recursively if target is a directory. In case of ./... execution will be recursively applied by default. Optional parameter.",
	)

	isUseCache = flag.Bool(
		useCacheArg,
		false,
		"Use cache to improve performance. Optional parameter.",
	)

	shouldApplyToGeneratedFiles = flag.Bool(
		applyToGeneratedFiles,
		false,
		"Apply imports sorting and formatting(if the option is set) to generated files. Generated file is a file with first comment which starts with comment '// Code generated'. Optional parameter.",
	)

	shouldShowVersion = flag.Bool(
		versionArg,
		false,
		"Show version information",
	)

	shouldShowVersionOnly = flag.Bool(
		versionOnlyArg,
		false,
		"Show only the version string",
	)

}

func printUsage() {
	if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0]); err != nil {
		log.Fatalf("failed to print usage: %s", err)
	}

	flag.PrintDefaults()
}

func getBuildInfo() *debug.BuildInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return bi
}

func getMyModuleInfo(bi *debug.BuildInfo) (*debug.Module, error) {
	if bi == nil {
		return nil, errors.New("no build info available")
	}
	// depending on the context in which we are called, the main module may not be set
	if bi.Main.Path != "" {
		return &bi.Main, nil
	}
	// if the main module is not set, we need to find the dep that contains our module
	for _, m := range bi.Deps {
		if modulePathMatcher.MatchString(m.Path) {
			return m, nil
		}
	}
	return nil, errors.New("no matching module found in build info")
}

func printVersion() {
	if Tag != "" {
		fmt.Printf(
			"version: %s\nbuilt with: %s\ntag: %s\ncommit: %s\nsource: %s\n",
			strings.TrimPrefix(Tag, "v"),
			GoVersion,
			Tag,
			Commit,
			SourceURL,
		)
		return
	}
	bi := getBuildInfo()
	myModule, err := getMyModuleInfo(bi)
	if err != nil {
		log.Fatalf("failed to get my module info: %s", err)
	}
	fmt.Printf(
		"version: %s\nbuilt with: %s\ntag: %s\ncommit: %s\nsource: %s\n",
		strings.TrimPrefix(myModule.Version, "v"),
		bi.GoVersion,
		myModule.Version,
		"n/a",
		myModule.Path,
	)
}

func printVersionOnly() {
	if Tag != "" {
		fmt.Println(strings.TrimPrefix(Tag, "v"))
		return
	}
	bi := getBuildInfo()
	myModule, err := getMyModuleInfo(bi)
	if err != nil {
		log.Fatalf("failed to get my module info: %s", err)
	}
	fmt.Println(strings.TrimPrefix(myModule.Version, "v"))
}

func main() {
	deprecatedMessagesCh := make(chan string, 10)
	flag.Parse()

	if shouldShowVersionOnly != nil && *shouldShowVersionOnly {
		printVersionOnly()
		return
	}

	if shouldShowVersion != nil && *shouldShowVersion {
		printVersion()
		return
	}

	originPath := flag.Arg(0)

	if filePath != "" {
		deprecatedMessagesCh <- fmt.Sprintf("-%s is deprecated. Put file name as last argument to the command(Example: goimports-reviser -rm-unused -set-alias -format goimports-reviser/main.go)", filePathArg)
		originPath = filePath
	}

	if originPath == "" {
		originPath = reviser.StandardInput
	}

	if err := validateRequiredParam(originPath); err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	var options reviser.SourceFileOptions
	if shouldRemoveUnusedImports != nil && *shouldRemoveUnusedImports {
		options = append(options, reviser.WithRemovingUnusedImports)
	}

	if shouldSetAlias != nil && *shouldSetAlias {
		options = append(options, reviser.WithUsingAliasForVersionSuffix)
	}

	if shouldFormat != nil && *shouldFormat {
		options = append(options, reviser.WithCodeFormatting)
	}

	if shouldApplyToGeneratedFiles == nil || !*shouldApplyToGeneratedFiles {
		options = append(options, reviser.WithSkipGeneratedFile)
	}

	if localPkgPrefixes != "" {
		if companyPkgPrefixes != "" {
			companyPkgPrefixes = localPkgPrefixes
		}
		deprecatedMessagesCh <- fmt.Sprintf(`-%s is deprecated and will be removed soon. Use -%s instead.`, localArg, companyPkgPrefixesArg)
	}

	if companyPkgPrefixes != "" {
		options = append(options, reviser.WithCompanyPackagePrefixes(companyPkgPrefixes))
	}

	if importsOrder != "" {
		order, err := reviser.StringToImportsOrders(importsOrder)
		if err != nil {
			fmt.Printf("%s\n\n", err)
			printUsage()
			os.Exit(1)
		}
		options = append(options, reviser.WithImportsOrder(order))
	}

	originProjectName, err := helper.DetermineProjectName(projectName, originPath, helper.OSGetwdOption)
	if err != nil {
		fmt.Printf("%s\n\n", err)
		printUsage()
		os.Exit(1)
	}

	close(deprecatedMessagesCh)

	if _, ok := reviser.IsDir(originPath); ok {
		if *listFileName {
			unformattedFiles, err := reviser.NewSourceDir(originProjectName, originPath, *isRecursive, excludes).Find(options...)
			if err != nil {
				log.Fatalf("Failed to find unformatted files %s: %+v\n", originPath, err)
			}
			fmt.Printf("%s\n", unformattedFiles.String())
			return
		}
		err := reviser.NewSourceDir(originProjectName, originPath, *isRecursive, excludes).Fix(options...)
		if err != nil {
			log.Fatalf("Failed to fix directory %s: %+v\n", originPath, err)
		}
		return
	}

	if originPath != reviser.StandardInput {
		originPath, err = filepath.Abs(originPath)
		if err != nil {
			log.Fatalf("Failed to get abs path: %+v\n", err)
		}
	}

	var formattedOutput []byte
	var hasChange bool
	if *isUseCache {
		hash := md5.Sum([]byte(originPath))

		u, err := user.Current()
		if err != nil {
			log.Fatalf("Failed to get current user: %+v\n", err)
		}
		cacheDir := path.Join(u.HomeDir, ".cache", "goimports-reviser")
		if err = os.MkdirAll(cacheDir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create cache directory: %+v\n", err)
		}
		cacheFile := path.Join(cacheDir, hex.EncodeToString(hash[:]))

		var cacheContent, fileContent []byte
		if cacheContent, err = os.ReadFile(cacheFile); err == nil {
			// compare file content hash
			var fileHashHex string
			if fileContent, err = os.ReadFile(originPath); err == nil {
				fileHash := md5.Sum(fileContent)
				fileHashHex = hex.EncodeToString(fileHash[:])
			}
			if string(cacheContent) == fileHashHex {
				// point to cache
				return
			}
		}
		formattedOutput, _, hasChange, err = reviser.NewSourceFile(originProjectName, originPath).Fix(options...)
		if err != nil {
			log.Fatalf("Failed to fix file: %+v\n", err)
		}
		fileHash := md5.Sum(formattedOutput)
		fileHashHex := hex.EncodeToString(fileHash[:])
		if fileInfo, err := os.Stat(cacheFile); err != nil || fileInfo.IsDir() {
			if _, err = os.Create(cacheFile); err != nil {
				log.Fatalf("Failed to create cache file: %+v\n", err)
			}
		}
		file, _ := os.OpenFile(cacheFile, os.O_RDWR, os.ModePerm)
		defer func() {
			_ = file.Close()
		}()
		if err = file.Truncate(0); err != nil {
			log.Fatalf("Failed file truncate: %+v\n", err)
		}
		if _, err = file.Seek(0, 0); err != nil {
			log.Fatalf("Failed file seek: %+v\n", err)
		}
		if _, err = file.WriteString(fileHashHex); err != nil {
			log.Fatalf("Failed to write file hash: %+v\n", err)
		}
	} else {
		formattedOutput, _, hasChange, err = reviser.NewSourceFile(originProjectName, originPath).Fix(options...)
		if err != nil {
			log.Fatalf("Failed to fix file: %+v\n", err)
		}
	}

	resultPostProcess(hasChange, deprecatedMessagesCh, originPath, formattedOutput)
}

func resultPostProcess(hasChange bool, deprecatedMessagesCh chan string, originFilePath string, formattedOutput []byte) {
	if !hasChange && *listFileName {
		printDeprecations(deprecatedMessagesCh)
		return
	}
	switch {
	case hasChange && *listFileName && output != "write":
		fmt.Println(originFilePath)
	case output == "stdout" || originFilePath == reviser.StandardInput:
		fmt.Print(string(formattedOutput))
	case output == "file" || output == "write":
		if !hasChange {
			printDeprecations(deprecatedMessagesCh)
			return
		}

		if err := os.WriteFile(originFilePath, formattedOutput, 0o644); err != nil {
			log.Fatalf("failed to write fixed result to file(%s): %+v\n", originFilePath, err)
		}
		if *listFileName {
			fmt.Println(originFilePath)
		}
	default:
		log.Fatalf(`invalid output %q specified`, output)
	}

	if hasChange && *setExitStatus {
		os.Exit(1)
	}

	printDeprecations(deprecatedMessagesCh)
}

func validateRequiredParam(filePath string) error {
	if filePath == reviser.StandardInput {
		stat, _ := os.Stdin.Stat()
		if stat.Mode()&os.ModeNamedPipe == 0 {
			// no data on stdin
			return errors.New("no data on stdin")
		}
	}
	return nil
}

func printDeprecations(deprecatedMessagesCh chan string) {
	var hasDeprecations bool
	for deprecatedMessage := range deprecatedMessagesCh {
		hasDeprecations = true
		fmt.Printf("%s\n", deprecatedMessage)
	}
	if hasDeprecations {
		fmt.Printf("All changes to file are applied, but command-line syntax should be fixed\n")
		os.Exit(1)
	}
}
