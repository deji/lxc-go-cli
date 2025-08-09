package helpers

import (
	"os"
	"testing"
)

func TestParseImageString(t *testing.T) {
	tests := []struct {
		input   string
		distro  string
		release string
		arch    string
	}{
		{"ubuntu:24.04:amd64", "ubuntu", "24.04", "amd64"},
		{"debian:12:arm64", "debian", "12", "arm64"},
		{"alpine", "alpine", "24.04", "amd64"},
		{"", "ubuntu", "24.04", "amd64"},
		{"centos:7", "centos", "7", "amd64"},
		{"ubuntu:", "ubuntu", "24.04", "amd64"},
		{"ubuntu:24.04:", "ubuntu", "24.04", "amd64"},
		{":24.04:amd64", "ubuntu", "24.04", "amd64"},
		{":24.04:", "ubuntu", "24.04", "amd64"},
		{":", "ubuntu", "24.04", "amd64"},
	}
	for _, tt := range tests {
		d, r, a := ParseImageString(tt.input)
		if d != tt.distro || r != tt.release || a != tt.arch {
			t.Errorf("ParseImageString(%q) = (%q, %q, %q), want (%q, %q, %q)", tt.input, d, r, a, tt.distro, tt.release, tt.arch)
		}
	}
}

func TestParseBtrfsPoolsFromJSON(t *testing.T) {
	// Test valid JSON with Btrfs pools
	jsonOutput := `[{"name":"default","driver":"dir"},{"name":"docker200","driver":"btrfs"}]`
	pools := parseBtrfsPoolsFromJSON(jsonOutput)

	expected := []string{"docker200"}
	if len(pools) != len(expected) || pools[0] != expected[0] {
		t.Errorf("got %v, want %v", pools, expected)
	}
}

func TestParseBtrfsPoolsFromJSON_NoBtrfsPools(t *testing.T) {
	// Test valid JSON without Btrfs pools
	jsonOutput := `[{"name":"default","driver":"dir"},{"name":"zfs-pool","driver":"zfs"}]`
	pools := parseBtrfsPoolsFromJSON(jsonOutput)

	if len(pools) != 0 {
		t.Errorf("got %v, want empty slice", pools)
	}
}

func TestParseBtrfsPoolsFromJSON_InvalidJSON(t *testing.T) {
	// Test invalid JSON (should fallback to table parsing)
	jsonOutput := `invalid json`

	pools := parseBtrfsPoolsFromJSON(jsonOutput)

	// Should return empty slice when JSON parsing fails
	if len(pools) != 0 {
		t.Errorf("got %v, want empty slice", pools)
	}
}

func TestParseBtrfsPoolsFromJSON_EmptyArray(t *testing.T) {
	// Test empty JSON array
	jsonOutput := `[]`
	pools := parseBtrfsPoolsFromJSON(jsonOutput)

	if len(pools) != 0 {
		t.Errorf("got %v, want empty slice", pools)
	}
}

func TestParseBtrfsPoolsFromJSON_MalformedObject(t *testing.T) {
	// Test JSON with malformed objects
	jsonOutput := `[{"name":"default"},{"name":"docker200","driver":"btrfs"}]`
	pools := parseBtrfsPoolsFromJSON(jsonOutput)

	expected := []string{"docker200"}
	if len(pools) != len(expected) || pools[0] != expected[0] {
		t.Errorf("got %v, want %v", pools, expected)
	}
}

func TestParseBtrfsPoolsFromTable(t *testing.T) {
	// Test table format with Btrfs pools
	tableOutput := `+-----------+--------+------------------------------------------------+-------------+---------+---------+
|   NAME    | DRIVER |                     SOURCE                     | DESCRIPTION | USED BY |  STATE  |
+-----------+--------+------------------------------------------------+-------------+---------+---------+
| default   | dir    | /var/snap/lxd/common/lxd/storage-pools/default |             | 3       | CREATED |
+-----------+--------+------------------------------------------------+-------------+---------+---------+
| docker200 | btrfs  | /var/snap/lxd/common/lxd/disks/docker200.img   |             | 1       | CREATED |
+-----------+--------+------------------------------------------------+-------------+---------+---------+`

	pools := parseBtrfsPoolsFromTable(tableOutput)

	expected := []string{"docker200"}
	if len(pools) != len(expected) || pools[0] != expected[0] {
		t.Errorf("got %v, want %v", pools, expected)
	}
}

func TestParseBtrfsPoolsFromTable_NoBtrfsPools(t *testing.T) {
	// Test table format without Btrfs pools
	tableOutput := `+-----------+--------+------------------------------------------------+-------------+---------+---------+
|   NAME    | DRIVER |                     SOURCE                     | DESCRIPTION | USED BY |  STATE  |
+-----------+--------+------------------------------------------------+-------------+---------+---------+
| default   | dir    | /var/snap/lxd/common/lxd/storage-pools/default |             | 3       | CREATED |
+-----------+--------+------------------------------------------------+-------------+---------+---------+`

	pools := parseBtrfsPoolsFromTable(tableOutput)

	if len(pools) != 0 {
		t.Errorf("got %v, want empty slice", pools)
	}
}

func TestParseBtrfsPoolsFromTable_MalformedTable(t *testing.T) {
	// Test malformed table
	tableOutput := `invalid table format`

	pools := parseBtrfsPoolsFromTable(tableOutput)

	if len(pools) != 0 {
		t.Errorf("got %v, want empty slice", pools)
	}
}

func TestParseBtrfsPoolsFromTable_EmptyTable(t *testing.T) {
	// Test empty table
	tableOutput := ``

	pools := parseBtrfsPoolsFromTable(tableOutput)

	if len(pools) != 0 {
		t.Errorf("got %v, want empty slice", pools)
	}
}

func TestParseBtrfsPoolsFromTable_HeaderOnly(t *testing.T) {
	// Test table with only header
	tableOutput := `+-----------+--------+------------------------------------------------+-------------+---------+---------+
|   NAME    | DRIVER |                     SOURCE                     | DESCRIPTION | USED BY |  STATE  |
+-----------+--------+------------------------------------------------+-------------+---------+---------+`

	pools := parseBtrfsPoolsFromTable(tableOutput)

	if len(pools) != 0 {
		t.Errorf("got %v, want empty slice", pools)
	}
}

func TestParseBtrfsPoolsFromTable_MultipleBtrfsPools(t *testing.T) {
	// Test table with multiple Btrfs pools
	tableOutput := `+-----------+--------+------------------------------------------------+-------------+---------+---------+
|   NAME    | DRIVER |                     SOURCE                     | DESCRIPTION | USED BY |  STATE  |
+-----------+--------+------------------------------------------------+-------------+---------+---------+
| default   | dir    | /var/snap/lxd/common/lxd/storage-pools/default |             | 3       | CREATED |
+-----------+--------+------------------------------------------------+-------------+---------+---------+
| docker200 | btrfs  | /var/snap/lxd/common/lxd/disks/docker200.img   |             | 1       | CREATED |
+-----------+--------+------------------------------------------------+-------------+---------+---------+
| backup    | btrfs  | /var/snap/lxd/common/lxd/disks/backup.img      |             | 0       | CREATED |
+-----------+--------+------------------------------------------------+-------------+---------+---------+`

	pools := parseBtrfsPoolsFromTable(tableOutput)

	expected := []string{"docker200", "backup"}
	if len(pools) != len(expected) {
		t.Errorf("got %v, want %v", pools, expected)
	}
	// Check that both expected pools are present (order doesn't matter)
	for _, expectedPool := range expected {
		found := false
		for _, pool := range pools {
			if pool == expectedPool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected pool %s not found in %v", expectedPool, pools)
		}
	}
}

func TestParseBtrfsPoolsFromTable_EdgeCases(t *testing.T) {
	// Test edge cases in table parsing
	tests := []struct {
		name     string
		table    string
		expected []string
	}{
		{
			name: "empty pool name",
			table: `|   NAME    | DRIVER | SOURCE |
|           | btrfs  | /path   |`,
			expected: []string{},
		},
		{
			name: "case insensitive btrfs",
			table: `|   NAME    | DRIVER | SOURCE |
| docker200 | BTRFS  | /path   |`,
			expected: []string{},
		},
		{
			name: "btrfs in source field",
			table: `|   NAME    | DRIVER | SOURCE                    |
| docker200 | dir    | /path/to/btrfs/mountpoint |`,
			expected: []string{},
		},
		{
			name: "incomplete line with btrfs",
			table: `|   NAME    | DRIVER | SOURCE |
| docker200 | btrfs  |`,
			expected: []string{"docker200"},
		},
		{
			name: "incomplete line without btrfs",
			table: `|   NAME    | DRIVER | SOURCE |
| docker200 | dir    |`,
			expected: []string{},
		},
		{
			name:     "single field line",
			table:    `| docker200 |`,
			expected: []string{},
		},
		{
			name:     "empty line with pipes",
			table:    `| | | |`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pools := parseBtrfsPoolsFromTable(tt.table)
			if len(pools) != len(tt.expected) {
				t.Errorf("got %v, want %v", pools, tt.expected)
			}
		})
	}
}

// TestHelperProcess is not a real test. It's used as a helper for exec.Command mocking.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("ERR") == "1" {
		os.Exit(1)
	}
	os.Stdout.WriteString(os.Getenv("STDOUT"))
	os.Exit(0)
}
