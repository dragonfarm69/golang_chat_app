import { useState } from "react"
import LoginPage from "./LoginPage"
import RegisterPage from "./RegisterPage"

function MainAuthenticationPage () {
    const [isRegistering, setIsRegistering] = useState<boolean>(false)
    return (
        <>
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
        </>
    )
}

export default MainAuthenticationPage