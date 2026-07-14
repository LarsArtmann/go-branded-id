import type { StepCard, ComparisonItem, UseCase, ComparisonMatrix } from "./types";

export const steps: StepCard[] = [
  {
    step: "1",
    stepColor: "accent",
    title: "Define Brand",
    desc: "Create an empty struct as a phantom brand type.",
    code: "type UserBrand struct{}",
  },
  {
    step: "2",
    stepColor: "accent",
    title: "Create ID Type",
    desc: "Alias ID with your brand and value type.",
    code: "type UserID = id.ID[UserBrand, string]",
  },
  {
    step: "3",
    stepColor: "violet",
    title: "Use in Functions",
    desc: "Function signatures enforce the brand at compile time.",
    code: "func GetUser(id UserID) error",
  },
  {
    step: "4",
    stepColor: "violet",
    title: "Compile-Time Safety",
    desc: "Passing the wrong ID type is a compile error, not a runtime bug.",
    code: "GetOrder(userID) // won't compile",
  },
];

export const comparisons: ComparisonItem[] = [
  {
    variant: "Plain types",
    accent: false,
    pros: ["Zero dependencies", "Familiar to all Go developers"],
    cons: [
      "No type safety between IDs",
      "Easy to swap userID and orderID",
      "Bugs only caught at runtime",
      "No debug context in strings",
    ],
  },
  {
    variant: "go-branded-id",
    accent: true,
    pros: [
      "Compile-time type safety",
      "Zero allocations on core ops",
      "Full serialization (JSON, SQL, Text, Binary, Gob)",
      "Named brands for debug visibility",
      "Stdlib-only, no dependencies",
    ],
    cons: [],
  },
];

export const comparisonMatrix: ComparisonMatrix = {
  columns: ["Plain types", "go-branded-id"],
  rows: [
    { feature: "Compile-time type safety", values: ["no", "yes"] },
    { feature: "Zero-allocation core ops", values: ["yes", "yes"] },
    { feature: "Debug-visible IDs", values: ["no", "yes"] },
    { feature: "JSON serialization", values: ["manual", "built-in"] },
    { feature: "SQL Scan / Value", values: ["manual", "built-in"] },
    { feature: "Brand-aware validation", values: ["no", "yes"] },
    { feature: "Third-party dependencies", values: ["0", "0"] },
  ],
};

export const useCases: UseCase[] = [
  {
    title: "REST APIs",
    desc: "Route params and request bodies carry typed IDs that can't be mixed",
    icon: "globe",
  },
  {
    title: "Database Models",
    desc: "Scan directly from SQL rows. Zero values serialize to null",
    icon: "database",
  },
  {
    title: "Domain-Driven Design",
    desc: "Strong domain types that make impossible states unrepresentable",
    icon: "cubes",
  },
];
