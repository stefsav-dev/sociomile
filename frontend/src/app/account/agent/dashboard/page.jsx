import ChatWidgetAgent from "@/components/utils/ChatWidgetAgent";
import { NavbarSignInOut } from "@/components/utils/NavbarSignInOut";

export default function AgentDashboard() {
    return (
        <>
            <NavbarSignInOut/>
            <main className="container mx-auto px-4 py-8">
                <h1 className="text-3xl font-bold">Halaman Agent</h1>
            </main>
            <ChatWidgetAgent/>
        </>
    )
}