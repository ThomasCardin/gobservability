package types

type PidDetails struct {
	// From /proc/{PID}/stat - process basics
	Name             string `json:"name"`              // Filename of executable
	State            string `json:"state"`             // Process state
	Priority         int    `json:"priority"`          // Priority level
	Nice             int    `json:"nice"`              // Nice level
	Threads          int    `json:"threads"`           // Number of threads
	StartTime        uint64 `json:"start_time"`        // Start time after boot
	RealtimePriority int    `json:"realtime_priority"` // Realtime priority
	CUTime           uint64 `json:"cutime"`            // Children user time
	CSTime           uint64 `json:"cstime"`            // Children system time
	TaskCPU          int    `json:"task_cpu"`          // Which CPU the task is scheduled on

	// From /proc/{PID}/status - security and scheduling
	KThread                   int    `json:"kthread"`                     // Kernel thread flag (1=yes, 0=no)
	Seccomp                   int    `json:"seccomp"`                     // Seccomp mode
	SpeculationIndirectBranch string `json:"speculation_indirect_branch"` // Indirect branch speculation
	CpusAllowedList           string `json:"cpus_allowed_list"`           // CPUs this process may run on
	MemsAllowedList           string `json:"mems_allowed_list"`           // Memory nodes allowed
	VoluntaryCtxtSwitches     uint64 `json:"voluntary_ctxt_switches"`     // Voluntary context switches
	NonvoluntaryCtxtSwitches  uint64 `json:"nonvoluntary_ctxt_switches"`  // Non-voluntary context switches
	VmPeak                    uint64 `json:"vm_peak"`                     // Peak virtual memory size (KB)
	VmLck                     uint64 `json:"vm_lck"`                      // Locked memory size (KB)
	VmPin                     uint64 `json:"vm_pin"`                      // Pinned memory size (KB)

	// From /proc/{PID}/net - network
	PacketsReceived    uint64 `json:"packets_received"`
	PacketsTransmitted uint64 `json:"packets_transmitted"`
	ErrorsReceived     uint64 `json:"errors_received"`
	ErrorsTransmitted  uint64 `json:"errors_transmitted"`

	// From /proc/{PID}/io
	CancelledWrites uint64 `json:"cancelled_writes"`

	// Additional process information
	Cmdline string   `json:"cmdline"`  // Command line with arguments
	Stack   []string `json:"stack"`    // Kernel stack trace
	OpenFDs int      `json:"open_fds"` // Number of open file descriptors
	MaxFDs  uint64   `json:"max_fds"`  // Maximum file descriptors allowed
	Cgroup  []string `json:"cgroup"`   // Control groups
	VmData  uint64   `json:"vm_data"`  // Data segment size (KB)
	VmStk   uint64   `json:"vm_stk"`   // Stack segment size (KB)
	VmExe   uint64   `json:"vm_exe"`   // Text segment size (KB)
	VmLib   uint64   `json:"vm_lib"`   // Shared library size (KB)
	VmSwap  uint64   `json:"vm_swap"`  // Swap usage (KB)
}
