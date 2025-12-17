import "../../App.css";
import { useState } from "react";
import { invoke } from "@tauri-apps/api/core";

function RegisterPage() {
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
        <div className="app-container" style={{backgroundColor: "white"}}>
            <form className="room-form" onSubmit={(e) => {e.preventDefault(); handleSubmit(e)}}> 
                <input 
                    placeholder="Input first name" 
                    type="text" 
                    value={firstName} 
                    onChange={(e) => setFirstName(e.target.value)}>
                </input>
                <input 
                    placeholder="Input last name" 
                    type="text" 
                    value={lastName} 
                    onChange={(e) => setLastName(e.target.value)}>
                </input>
                <input 
                    placeholder="Input email" 
                    type="text" 
                    value={email} 
                    onChange={(e) => setEmail(e.target.value)}>
                </input>
                <input 
                    placeholder="Input password" 
                    type="password" 
                    value={password} 
                    onChange={(e) => setUserPassword(e.target.value)}>
                </input>
                <input 
                    placeholder="Reinput password" 
                    type="password" 
                    value={passwordCheck} 
                    onChange={(e) => setUserPasswordCheck(e.target.value)}>
                </input>
                <button className="join-room-button" type="submit">Register</button>
            </form>
        </div>
    )
}

export default RegisterPage