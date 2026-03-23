import "../../App.css";
import { useState } from "react";
import { commands, RegisterPayload } from "../../bindings";

function validateEmail(email: string): string | null {
  if (!email) return null;
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return re.test(email) ? null : "Please enter a valid email address.";
}

interface PasswordStrength {
  errors: string[];
  strength: "weak" | "medium" | "strong" | null;
}

function validatePassword(password: string): PasswordStrength {
  if (!password) return { errors: [], strength: null };
  const errors: string[] = [];
  if (password.length < 12) errors.push("At least 12 characters");
  if (!/[A-Z]/.test(password)) errors.push("At least one uppercase letter");
  if (!/[0-9]/.test(password)) errors.push("At least one number");
  if (!/[^A-Za-z0-9]/.test(password))
    errors.push("At least one special character");

  const passed = 4 - errors.length;
  const strength = passed <= 1 ? "weak" : passed <= 3 ? "medium" : "strong";

  return { errors, strength };
}

function RegisterPage({ isRegistering }: { isRegistering: () => void }) {
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setUserPassword] = useState("");
  const [passwordCheck, setUserPasswordCheck] = useState("");
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const emailError = validateEmail(email);
  const { errors: pwErrors, strength: pwStrength } = validatePassword(password);
  const passwordMismatch =
    touched.passwordCheck &&
    passwordCheck.length > 0 &&
    password !== passwordCheck;

  const isFormValid =
    firstName.trim() &&
    lastName.trim() &&
    !emailError &&
    email.length > 0 &&
    pwErrors.length === 0 &&
    password.length >= 12 &&
    password === passwordCheck;

  function touch(field: string) {
    setTouched((prev) => ({ ...prev, [field]: true }));
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setTouched({ email: true, password: true, passwordCheck: true });

    if (!isFormValid) return;

    try {
      const payload: RegisterPayload = {
        email: email,
        first_name: firstName,
        last_name: lastName,
        password: password,
      };
      const result = await commands.register(payload);

      console.log("Registering: ", result);
    } catch (error) {
      alert(`Registration failed: ${error}`);
    }
  }

  const strengthColor =
    pwStrength === "strong"
      ? "#22c55e"
      : pwStrength === "medium"
        ? "#f59e0b"
        : "#ef4444";

  return (
    <>
      <form className="register-form" onSubmit={handleSubmit} noValidate>
        {/* Name row */}
        <div className="form-row">
          <div className="form-field">
            <label htmlFor="firstName">First Name</label>
            <input
              id="firstName"
              placeholder="Enter first name"
              type="text"
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              required
            />
          </div>
          <div className="form-field">
            <label htmlFor="lastName">Last Name</label>
            <input
              id="lastName"
              placeholder="Enter last name"
              type="text"
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              required
            />
          </div>
        </div>

        {/* Email */}
        <div className="form-field">
          <label htmlFor="email">Email</label>
          <input
            id="email"
            placeholder="Enter your email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            onBlur={() => touch("email")}
            style={
              touched.email && emailError
                ? { borderColor: "#ef4444" }
                : touched.email && !emailError && email
                  ? { borderColor: "#22c55e" }
                  : {}
            }
            required
          />
          {touched.email && emailError && (
            <span className="field-error">{emailError}</span>
          )}
        </div>

        {/* Password */}
        <div className="form-field">
          <label htmlFor="password">Password</label>
          <input
            id="password"
            placeholder="Enter password"
            type="password"
            value={password}
            onChange={(e) => setUserPassword(e.target.value)}
            onBlur={() => touch("password")}
            style={
              touched.password && pwErrors.length > 0
                ? { borderColor: "#ef4444" }
                : touched.password && pwErrors.length === 0 && password
                  ? { borderColor: "#22c55e" }
                  : {}
            }
            required
          />

          {/* Strength bar */}
          {password.length > 0 && (
            <div className="strength-bar-wrapper">
              <div
                className="strength-bar-fill"
                style={{
                  width:
                    pwStrength === "strong"
                      ? "100%"
                      : pwStrength === "medium"
                        ? "60%"
                        : "25%",
                  backgroundColor: strengthColor,
                }}
              />
            </div>
          )}

          {/* Inline checklist */}
          {(touched.password || password.length > 0) && pwErrors.length > 0 && (
            <ul className="password-checklist">
              {[
                { label: "At least 12 characters", ok: password.length >= 12 },
                { label: "Uppercase letter", ok: /[A-Z]/.test(password) },
                { label: "Number", ok: /[0-9]/.test(password) },
                {
                  label: "Special character",
                  ok: /[^A-Za-z0-9]/.test(password),
                },
              ].map(({ label, ok }) => (
                <li key={label} className={ok ? "check-ok" : "check-fail"}>
                  <span>{ok ? "✓" : "✗"}</span> {label}
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Confirm password */}
        <div className="form-field">
          <label htmlFor="passwordCheck">Confirm Password</label>
          <input
            id="passwordCheck"
            placeholder="Re-enter password"
            type="password"
            value={passwordCheck}
            onChange={(e) => setUserPasswordCheck(e.target.value)}
            onBlur={() => touch("passwordCheck")}
            style={
              passwordMismatch
                ? { borderColor: "#ef4444" }
                : touched.passwordCheck && !passwordMismatch && passwordCheck
                  ? { borderColor: "#22c55e" }
                  : {}
            }
            required
          />
          {passwordMismatch && (
            <span className="field-error">Passwords do not match.</span>
          )}
        </div>

        <button
          className="login-button"
          type="submit"
          disabled={!isFormValid}
          style={isFormValid ? {} : { opacity: 0.5, cursor: "not-allowed" }}
        >
          <span>Create Account</span>
        </button>
      </form>

      <div className="login-footer">
        <p>
          Already have an account?{" "}
          <a className="register-link" onClick={() => isRegistering()}>
            Login
          </a>
        </p>
      </div>
    </>
  );
}

export default RegisterPage;
