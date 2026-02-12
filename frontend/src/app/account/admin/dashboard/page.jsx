import { NavbarSignInOut } from "@/components/utils/NavbarSignInOut";

export default function AdminDashboard() {
    return (
        <>
            <NavbarSignInOut/>
            <main className="container mx-auto px-4 py-8">
                <h1 className="text-3xl font-bold">Hallaman Admin</h1>
            </main>
        </>
    )
}