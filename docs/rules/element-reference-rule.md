# element-reference-rule

Checks that layout element references resolve to defined panels, and that every panel is placed in the layout.

# Description

A v2 dashboard defines its panels in `spec.elements` and places them through `ElementReference` nodes in the layout tree. The v2 CUE schema does not check those references, so a renamed or deleted panel leaves a stale reference in the layout, and the dashboard renders incomplete or fails at render time.

The rule walks the whole spec and collects every `ElementReference` name. It reports:

- an error for each referenced name with no definition in `spec.elements` (a broken layout),
- a warning for each defined panel that no layout node references (an orphaned panel that will never render).

The rule reports the names sorted and deduplicated. It only runs on the kubernetes v2 dashboards (`apiVersion` v2alpha1, v2beta1, and later); the classic dashboards have no layout references.

# Best Practice

Every panel defined in `spec.elements` should be placed in the layout exactly once, and every layout reference should name a defined panel.

# Possible exceptions

A panel can be intentionally kept out of the layout while a feature is in development. In this case you may wish to create a lint exclusion for this rule.
