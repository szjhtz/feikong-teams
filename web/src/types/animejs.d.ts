declare module "animejs" {
  interface AnimeParams {
    targets?: unknown;
    opacity?: unknown;
    translateY?: unknown;
    duration?: number;
    easing?: string;
    [key: string]: unknown;
  }

  export default function anime(params: AnimeParams): unknown;
}
