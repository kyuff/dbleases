package lease

import (
	"testing"

	"github.com/kyuff/dbleases/internal/assert"
)

func TestAnalyzeLeases(t *testing.T) {
	t.Run("should take full ring as only lease", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "my-client", Name: "lease-a", Value: 1, Status: Leased},
			}
			expectApprove []Info
		)

		// act
		got := sut.Analyze("my-client", 4)

		// assert
		assert.EqualSlice(t, []int{0, 1, 2, 3}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should approve as only lease", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "my-client", Name: "lease-a", Value: 1, Status: Pending},
			}
			expectApprove = []Info{
				{ClientID: "my-client", Name: "lease-a", Value: 1, Status: Pending},
			}
		)

		// act
		got := sut.Analyze("my-client", 4)

		// assert
		assert.EqualSlice(t, []int{}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should respect next lease in ring", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "my-client", Name: "lease-a", Value: 2, Status: Leased},
				{ClientID: "client-1", Name: "lease-a", Value: 4, Status: Leased},
			}
			expectApprove []Info
		)

		// act
		got := sut.Analyze("my-client", 5)

		// assert
		assert.EqualSlice(t, []int{2, 3}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should approve next client in ring", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "my-client", Name: "lease-a", Value: 2, Status: Leased},
				{ClientID: "client-1", Name: "lease-a", Value: 4, Status: Pending},
			}
			expectApprove = []Info{
				{ClientID: "client-1", Name: "lease-a", Value: 4, Status: Pending},
			}
		)

		// act
		got := sut.Analyze("my-client", 5)

		// assert
		assert.EqualSlice(t, []int{2, 3}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should approve first pending client", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "my-client", Name: "lease-a", Value: 2, Status: Pending},
				{ClientID: "client-1", Name: "lease-a", Value: 4, Status: Pending},
			}
			expectApprove = []Info{
				{ClientID: "my-client", Name: "lease-a", Value: 2, Status: Pending},
			}
		)

		// act
		got := sut.Analyze("my-client", 5)

		// assert
		assert.EqualSlice(t, []int{}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should let first client approve", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "my-client", Name: "lease-a", Value: 4, Status: Pending},
				{ClientID: "client-1", Name: "lease-a", Value: 2, Status: Pending},
			}
			expectApprove []Info
		)

		// act
		got := sut.Analyze("my-client", 5)

		// assert
		assert.EqualSlice(t, []int{}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should calculate ring that passes the number end", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 2, Status: Leased},
				{ClientID: "my-client", Name: "lease-a", Value: 4, Status: Leased},
			}
			expectApprove []Info
		)

		// act
		got := sut.Analyze("my-client", 5)

		// assert
		assert.EqualSlice(t, []int{0, 1, 4}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should respect a larger ring", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 1, Status: Leased},
				{ClientID: "client-2", Name: "lease-a", Value: 4, Status: Leased},
				{ClientID: "my-client", Name: "lease-a", Value: 6, Status: Leased},
				{ClientID: "client-3", Name: "lease-a", Value: 8, Status: Leased},
			}
			expectApprove []Info
		)

		// act
		got := sut.Analyze("my-client", 10)

		// assert
		assert.EqualSlice(t, []int{6, 7}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should recognize a client with multiple spots in the ring", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "my-client", Name: "lease-a", Value: 1, Status: Leased},
				{ClientID: "client-2", Name: "lease-a", Value: 4, Status: Leased},
				{ClientID: "my-client", Name: "lease-a", Value: 6, Status: Leased},
				{ClientID: "client-3", Name: "lease-a", Value: 8, Status: Leased},
			}
			expectApprove []Info
		)

		// act
		got := sut.Analyze("my-client", 10)

		// assert
		assert.EqualSlice(t, []int{1, 2, 3, 6, 7}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should support when a client is both at the start and end", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "my-client", Name: "lease-a", Value: 1, Status: Leased},
				{ClientID: "client-1", Name: "lease-a", Value: 4, Status: Leased},
				{ClientID: "client-2", Name: "lease-a", Value: 6, Status: Leased},
				{ClientID: "my-client", Name: "lease-a", Value: 8, Status: Leased},
			}
			expectApprove []Info
		)

		// act
		got := sut.Analyze("my-client", 10)

		// assert
		assert.EqualSlice(t, []int{0, 1, 2, 3, 8, 9}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should approve next in the ring that is pending", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 1, Status: Leased},
				{ClientID: "my-client", Name: "lease-a", Value: 4, Status: Leased},
				{ClientID: "client-2", Name: "lease-a", Value: 6, Status: Pending},
				{ClientID: "client-3", Name: "lease-a", Value: 8, Status: Leased},
			}
			expectApprove = []Info{
				{ClientID: "client-2", Name: "lease-a", Value: 6, Status: Pending},
			}
		)

		// act
		got := sut.Analyze("my-client", 10)

		// assert
		assert.EqualSlice(t, []int{4, 5}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should ignore client that is before in the ring and pending", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 1, Status: Leased},
				{ClientID: "client-2", Name: "lease-a", Value: 4, Status: Pending},
				{ClientID: "my-client", Name: "lease-a", Value: 6, Status: Leased},
				{ClientID: "client-3", Name: "lease-a", Value: 8, Status: Leased},
			}
			expectApprove []Info
		)

		// act
		got := sut.Analyze("my-client", 10)

		// assert
		assert.EqualSlice(t, []int{6, 7}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should approve pending that is located after number end in the ring", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 1, Status: Pending},
				{ClientID: "client-2", Name: "lease-a", Value: 4, Status: Leased},
				{ClientID: "client-3", Name: "lease-a", Value: 6, Status: Leased},
				{ClientID: "my-client", Name: "lease-a", Value: 8, Status: Leased},
			}
			expectApprove = []Info{
				{ClientID: "client-1", Name: "lease-a", Value: 1, Status: Pending},
			}
		)

		// act
		got := sut.Analyze("my-client", 10)

		// assert
		assert.EqualSlice(t, []int{0, 8, 9}, got.Values)
		assert.EqualSlice(t, expectApprove, got.Approvals)
	})

	t.Run("should request balanced lease expansion", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 5, Status: Leased},
				{ClientID: "my-client", Name: "lease-a", Value: 13, Status: Leased},
				{ClientID: "client-2", Name: "lease-a", Value: 15, Status: Leased},
			}
			expectBalance = Request{
				ClientID: "my-client", LeaseName: "lease-a", Value: 56, Status: Pending,
			}
		)

		// act
		got := sut.Analyze("my-client", 100)

		// assert
		if assert.NotNil(t, got.Balance) {
			assert.Equal(t, expectBalance, *got.Balance)
		}
	})

	t.Run("should not request lease expansion as other client has fewer values", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 5, Status: Leased},
				{ClientID: "my-client", Name: "lease-a", Value: 9, Status: Leased},
				{ClientID: "client-2", Name: "lease-a", Value: 15, Status: Leased},
			}
		)

		// act
		got := sut.Analyze("my-client", 100)

		// assert
		assert.Nil(t, got.Balance)
	})

	t.Run("should request lease expansion", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 2, Status: Leased},  // 8/11
				{ClientID: "my-client", Name: "lease-a", Value: 0, Status: Leased}, // 2/11
			}
			expectBalance = Request{
				ClientID: "my-client", LeaseName: "lease-a", Value: 8, Status: Pending,
			}
		)

		// act
		got := sut.Analyze("my-client", 11)

		// assert
		if assert.NotNil(t, got.Balance) {
			assert.Equal(t, expectBalance, *got.Balance)
		}
	})

	t.Run("should not request lease expansion", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 0, Status: Leased},
				{ClientID: "client-2", Name: "lease-a", Value: 1, Status: Leased},
			}
		)

		// act
		got := sut.Analyze("my-client", 2)

		// assert
		assert.Nil(t, got.Balance)
	})

	t.Run("should request lease to enter the ring", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 0, Status: Leased},
				{ClientID: "client-2", Name: "lease-a", Value: 1, Status: Leased},
			}
			expectBalance = Request{
				ClientID: "my-client", LeaseName: "lease-a", Value: 9, Status: Pending,
			}
		)

		// act
		got := sut.Analyze("my-client", 10)

		// assert
		if assert.NotNil(t, got.Balance) {
			assert.Equal(t, expectBalance, *got.Balance)
		}
	})

	t.Run("should not balance a booting system", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 0, Status: Pending},
			}
		)

		// act
		got := sut.Analyze("my-client", 10)

		// assert
		assert.Nil(t, got.Balance)
	})

	t.Run("should balance a started system", func(t *testing.T) {
		// arrange
		var (
			sut = Ring{
				{ClientID: "client-1", Name: "lease-a", Value: 0, Status: Leased},
			}
			expectBalance = Request{
				ClientID: "my-client", LeaseName: "lease-a", Value: 9, Status: Pending,
			}
		)

		// act
		got := sut.Analyze("my-client", 10)

		// assert
		if assert.NotNil(t, got.Balance) {
			assert.Equal(t, expectBalance, *got.Balance)
		}
	})
}
