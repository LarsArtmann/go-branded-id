import type { Feature } from "./types";

export const features: Feature[] = [
  {
    icon: "shield",
    title: "Compile-Time Type Safety",
    desc: "Phantom types make it impossible to pass a UserID where an OrderID is expected. The compiler catches the bug before your code runs.",
  },
  {
    icon: "lightning",
    title: "Zero Allocations",
    desc: "Core operations (NewID, Get, Equal, Compare, IsZero) allocate nothing. NewID runs in 0.4 nanoseconds. No GC pressure.",
  },
  {
    icon: "layers",
    title: "Full Serialization",
    desc: "JSON, SQL, Text (XML/TOML), Binary, and Gob all implemented. Zero values serialize to null. Works with database/sql out of the box.",
  },
  {
    icon: "tag",
    title: "Named Brands",
    desc: "Optional Name() method enables debug-visible IDs like User:abc123 and brand-aware validation errors. Opt-in, zero cost if unused.",
  },
  {
    icon: "code",
    title: "Any Comparable Type",
    desc: "ID[Brand, V comparable] works with strings, ints, and any comparable type. Full serialization for 12 built-in numeric and string types.",
  },
  {
    icon: "database",
    title: "Stdlib-Only",
    desc: "No third-party dependencies. Uses encoding/json/v2 from the Go standard library. Nothing to audit, nothing to break.",
  },
];
