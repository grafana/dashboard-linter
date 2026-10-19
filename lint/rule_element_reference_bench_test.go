package lint

import (
	"fmt"
	"strings"
	"testing"
)

// benchmarkElementReferenceDashboard builds a v2 dashboard with n element
// panels and the supplied layout JSON.
func benchmarkElementReferenceDashboard(n int, layout string) (Dashboard, error) {
	definitions := make([]string, 0, n)
	for i := 0; i < n; i++ {
		definitions = append(definitions, v2ElementReferencePanel(fmt.Sprintf("panel-%d", i), i+1))
	}

	raw := fmt.Sprintf(`{
		"apiVersion": "dashboard.grafana.app/v2beta1",
		"kind": "Dashboard",
		"spec": {
			"title": "element references",
			"elements": { %s },
			"layout": %s
		}
	}`, strings.Join(definitions, ", "), layout)

	return NewDashboard([]byte(raw))
}

// gridLayout returns a GridLayout with one GridLayoutItem per element.
func gridLayout(n int) string {
	items := make([]string, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, fmt.Sprintf(
			`{"kind":"GridLayoutItem","spec":{"element":{"kind":"ElementReference","name":"panel-%d"},"x":%d,"y":%d,"width":12,"height":8}}`,
			i, (i%2)*12, (i/2)*8))
	}
	return fmt.Sprintf(`{"kind":"GridLayout","spec":{"items":[%s]}}`, strings.Join(items, ","))
}

// deepRowsLayout returns a RowsLayout nested depth levels deep. Each level
// carries one ElementReference, so the walk must recurse to the full depth.
func deepRowsLayout(depth int) string {
	var inner string
	for i := depth - 1; i >= 0; i-- {
		ref := fmt.Sprintf(`{"kind":"ElementReference","name":"panel-%d"}`, i)
		if inner == "" {
			inner = fmt.Sprintf(`{"kind":"RowsLayoutRow","spec":{"title":"row-%d","elements":[%s]}}`, i, ref)
		} else {
			nested := fmt.Sprintf(`{"kind":"RowsLayout","spec":{"rows":[%s]}}`, inner)
			inner = fmt.Sprintf(`{"kind":"RowsLayoutRow","spec":{"title":"row-%d","elements":[%s,%s]}}`, i, ref, nested)
		}
	}
	return fmt.Sprintf(`{"kind":"RowsLayout","spec":{"rows":[%s]}}`, inner)
}

func BenchmarkElementReferenceRule(b *testing.B) {
	sizes := []int{10, 50, 200}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("grid-%d", n), func(b *testing.B) {
			d, err := benchmarkElementReferenceDashboard(n, gridLayout(n))
			if err != nil {
				b.Fatalf("building dashboard: %v", err)
			}

			rule := NewElementReferenceRule()
			b.SetBytes(int64(len(d.Spec)))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				rs := ResultSet{}
				rule.Lint(d, &rs)
			}
		})
	}

	b.Run("deep-rows-100", func(b *testing.B) {
		n := 100
		d, err := benchmarkElementReferenceDashboard(n, deepRowsLayout(n))
		if err != nil {
			b.Fatalf("building dashboard: %v", err)
		}

		rule := NewElementReferenceRule()
		b.SetBytes(int64(len(d.Spec)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			rs := ResultSet{}
			rule.Lint(d, &rs)
		}
	})
}
