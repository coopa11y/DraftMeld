import { forwardRef, type ButtonHTMLAttributes } from "react";
import { classNames } from "./classNames";

type ButtonVariant = "primary" | "secondary" | "danger" | "dangerText" | "neutral";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  fullWidth?: boolean;
  variant?: ButtonVariant;
}

const variantClasses: Record<ButtonVariant, string> = {
  primary: "primary-button",
  secondary: "secondary-button",
  danger: "danger-button",
  dangerText: "danger-text-button",
  neutral: "neutral-button",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { className, fullWidth = false, type = "button", variant = "secondary", ...props },
  ref,
) {
  return (
    <button
      className={classNames(variantClasses[variant], fullWidth && "full-width-button", className)}
      ref={ref}
      type={type}
      {...props}
    />
  );
});
