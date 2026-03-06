import { useEffect, useState } from "react"
import LoginPage from "./LoginPage"
import RegisterPage from "./RegisterPage"

function MainAuthenticationPage () {
    const [isRegistering, setIsRegistering] = useState<boolean>(false)

    return (
        <>
            <div className="authentication-main-container">
                <div className={"authentication-container" + " " + (isRegistering ? "is-registering" : "")}>
                    <div className="authentication-card">
                        <div className="authentication-header">
                            <img src={"/src/assets/app_icon.png"} className="app-icon"></img>
                            {
                                isRegistering ?
                                <div>
                                    <h1 className="authentication-title">Creating account</h1> 
                                    <p className="authentication-subtitle">We're excited to have you join us!</p>
                                </div>
                                :
                                <h1 className="authentication-title">Welcome back!</h1>
                            }
                        </div>
                        {
                            isRegistering ?
                            <RegisterPage
                                isRegistering={() => setIsRegistering(false)}
                            >

                            </RegisterPage> 
                            :
                            <LoginPage
                                isRegistering={() => setIsRegistering(true)}
                            ></LoginPage>
                        }
                    </div>
                </div>
            </div>
        </>
    )
}

export default MainAuthenticationPage