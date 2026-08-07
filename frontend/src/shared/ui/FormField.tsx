import { cloneElement, useId, type ReactElement, type ReactNode } from "react";
import { classNames } from "./classNames";

interface FormFieldProps {
  children: ReactElement<{ "aria-describedby"?: string; id?: string }>;
  className?: string;
  help?: ReactNode;
  label: ReactNode;
}

export function FormField({ children, className, help, label }: FormFieldProps) {
  const generatedId = useId();
  const controlId = children.props.id ?? generatedId;
  const helpId = help ? `${controlId}-help` : undefined;
  const describedBy = [children.props["aria-describedby"], helpId].filter(Boolean).join(" ") || undefined;

  return (
    <div className={classNames("form-field", className)}>
      <label className="form-field-label" htmlFor={controlId}>{label}</label>
      {cloneElement(children, { id: controlId, "aria-describedby": describedBy })}
      {help ? <span className="field-help" id={helpId}>{help}</span> : null}
    </div>
  );
}
