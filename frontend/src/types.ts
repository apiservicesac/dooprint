export type ToastType = "success" | "danger";

export type Notify = (message: string, type?: ToastType) => void;

export interface Step {
  title: string;
  desc: string;
  link?: string;
  linkLabel?: string;
  image?: string;
  codes?: string[];
}
