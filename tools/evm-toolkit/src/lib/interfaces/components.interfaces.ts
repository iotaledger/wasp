import type { ComponentType } from 'svelte';

export interface ITab {
  value: number;
  label: string;
  component: ComponentType;
}
