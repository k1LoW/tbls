// go-assets is a simple embedding asset generator and consumer library for go.
// The main use of the library is to generate and embed small in-memory file
// systems ready to be integrated in webservers or other services which have
// a small amount of assets used at runtime. This is great for being able to do
// single binary deployments with assets.
//
// The Generator type can be used to generate a go file containing an in-memory
// file tree from files and directories on disk. The file data can be optionally
// compressed using gzip to reduce file size. Afterwards, the generated file
// can be included into your application and the assets can be directly accessed
// without having to load them from disk. The generated assets variable is of
// type FileSystem and implements the os.FileInfo and http.FileSystem interfaces
// so that they can be directly used with http.FileHandler.
package assets
