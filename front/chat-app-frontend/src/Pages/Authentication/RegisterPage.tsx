import "../../App.css";
import { useState } from "react";
import { invoke } from "@tauri-apps/api/core";

function RegisterPage({isRegistering} : {isRegistering: () => void}) {
    const [firstName, setFirstName] = useState("");
    const [lastName, setLastName] = useState("");
    const [email, setEmail] = useState("");
    const [password, setUserPassword] = useState("");
    const [passwordCheck, setUserPasswordCheck] = useState("");

    async function handleSubmit(e: any) {
        e.preventDefault();

        if(password != passwordCheck) {
            alert("Password doesn't match!");
            return;
        }

        try {
            const result = await invoke("register_account", {
                firstName,
                lastName,
                email,
                password
            })
            console.log(result);
        } catch (error) {
            alert(`Registration failed: ${error}`);
        }
    }

    return (
        <>       
            <form className="register-form" onSubmit={(e) => {e.preventDefault(); handleSubmit(e)}}> 
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
                
                <div className="form-field">
                    <label htmlFor="email">Email</label>
                    <input 
                        id="email"
                        placeholder="Enter your email" 
                        type="email" 
                        value={email} 
                        onChange={(e) => setEmail(e.target.value)}
                        required
                    />
                </div>
                
                <div className="form-field">
                    <label htmlFor="password">Password</label>
                    <input 
                        id="password"
                        placeholder="Enter password" 
                        type="password" 
                        value={password} 
                        onChange={(e) => setUserPassword(e.target.value)}
                        required
                    />
                </div>
                
                <div className="form-field">
                    <label htmlFor="passwordCheck">Confirm Password</label>
                    <input 
                        id="passwordCheck"
                        placeholder="Re-enter password" 
                        type="password" 
                        value={passwordCheck} 
                        onChange={(e) => setUserPasswordCheck(e.target.value)}
                        required
                    />
                </div>
                
                <button className="login-button" type="submit">
                    <span>Create Account</span>
                </button>
            </form>

            <div className="login-footer">
                <p>Already have an account? 
                    <a className="register-link"
                    onClick={() => isRegistering()}
                    >Login</a>
                    </p>
            </div>
        </>
    )
}

export default RegisterPage