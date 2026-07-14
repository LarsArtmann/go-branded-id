export const featureIconKeys = [
  "shield",
  "lightning",
  "layers",
  "tag",
  "code",
  "database",
] as const;
export type FeatureIcon = (typeof featureIconKeys)[number];

export interface Feature {
  icon: FeatureIcon;
  title: string;
  desc: string;
}

export interface StepCard {
  step: string;
  stepColor: "accent" | "violet";
  title: string;
  desc: string;
  code?: string;
}

export type ComparisonVariant = "Plain types" | "go-branded-id";

export interface ComparisonItem {
  variant: ComparisonVariant;
  pros: string[];
  cons: string[];
  accent: boolean;
}

export type MatrixValue = "yes" | "no" | string;

export interface MatrixRow {
  feature: string;
  values: [MatrixValue, MatrixValue];
}

export interface ComparisonMatrix {
  columns: [ComparisonVariant, ComparisonVariant];
  rows: MatrixRow[];
}

export const useCaseIconKeys = ["globe", "database", "cubes"] as const;
export type UseCaseIcon = (typeof useCaseIconKeys)[number];

export interface UseCase {
  title: string;
  desc: string;
  icon: UseCaseIcon;
}

export const uiIconKeys = [
  "arrow-external",
  "arrow-right",
  "github",
  "menu",
  "close",
  "sun",
  "moon",
  "star",
] as const;
export type UIIcon = (typeof uiIconKeys)[number];

export type IconName = FeatureIcon | UseCaseIcon | UIIcon;
