package pkg

//import "syscall"
import "github.com/ricochet2200/go-disk-usage/du"

type Status struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

func DiskUsage(p string) (disk Status) {
	// fs := syscall.Statfs_t{}
	// err := syscall.Statfs(p, &fs)
	// if err != nil {
	// 	return
	// }

	usage := du.NewDiskUsage(p)

	//var KB = uint64(1024)

	free := usage.Free()
	size := usage.Size()
	//available := usage.Available()
	used := usage.Used()

	disk.All = size
	disk.Free = free
	disk.Used = used

	// disk.All = fs.Blocks * uint64(fs.Bsize)
	// disk.Free = fs.Bfree * uint64(fs.Bsize)
	// disk.Used = disk.All - disk.Free
	return
}
