"use client"

import { useState } from "react";
import { useRouter } from "next/navigation";

import { 
    Card,
    CardHeader,
    CardTitle,
    CardDescription,
    CardAction,
    CardContent,
    CardFooter
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import Link from "next/link";
import axios from "@/lib/axios"; 

export default function LoginPage() {
    const router = useRouter();
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");
    const [formData, setFormData] = useState({
        email: "",
        password: ""
    });

    const handleChange = (e) => {
        setFormData({
            ...formData,
            [e.target.id]: e.target.value
        });
        if (error) setError("");
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        

        if (!formData.email || !formData.password) {
            setError("Email dan password harus diisi");
            return;
        }

        setIsLoading(true);
        setError("");

        try {
            const response = await axios.post(`/auth/login`, {
                email: formData.email,
                password: formData.password
            });

            if (response.data.success) {
                const { access_token, token, user } = response.data.data;
                

                const tokenToSave = access_token || token;
                
                localStorage.setItem('token', tokenToSave);
                localStorage.setItem('user', JSON.stringify(user));
                

                axios.defaults.headers.common['Authorization'] = `Bearer ${tokenToSave}`;

                switch(user.role) {
                    case 'admin':
                        router.push('/account/admin/dashboard');
                        break;
                    case 'agent':
                        router.push('/account/agent/dashboard');
                        break;
                    case 'user':
                        router.push('/account/user/dashboard');
                        break;
                    default:
                        router.push('/dashboard');
                }
            }
        } catch (error) {
            console.error("Login error:", error);
            
            if (error.response) {

                if (error.response.status === 429) {
                    setError("Terlalu banyak percobaan login. Silakan tunggu 15 menit.");
                } else {
                    setError(error.response.data?.message || "Email atau password salah");
                }
            } else if (error.request) {
                setError("Tidak dapat terhubung ke server");
            } else {
                setError("Terjadi kesalahan, silakan coba lagi");
            }
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="flex min-h-screen items-center justify-center">
            <Card className="w-full max-w-sm">
                <CardHeader>
                    <CardTitle>Login to your account</CardTitle>
                    <CardDescription>
                        Enter your email below to login to your account
                    </CardDescription>
                    <CardAction>
                        <Button variant="link" asChild>
                            <Link href="/auth/register">Sign Up</Link>
                        </Button>
                    </CardAction>
                </CardHeader>
                <CardContent>
                    {error && (
                        <div className="mb-4 p-3 text-sm text-red-500 bg-red-50 rounded-md border border-red-200">
                            {error}
                        </div>
                    )}
                    <form onSubmit={handleSubmit}>
                        <div className="flex flex-col gap-6">
                            <div className="grid gap-2">
                                <Label htmlFor="email">Email</Label>
                                <Input
                                    id="email"
                                    type="email"
                                    placeholder="m@example.com"
                                    value={formData.email}
                                    onChange={handleChange}
                                    required
                                    disabled={isLoading}
                                />
                            </div>
                            <div className="grid gap-2">
                                <div className="flex items-center">
                                    <Label htmlFor="password">Password</Label>
                                </div>
                                <Input
                                    id="password"
                                    type="password"
                                    value={formData.password}
                                    onChange={handleChange}
                                    required
                                    disabled={isLoading}
                                />
                            </div>
                        </div>
                    </form>
                </CardContent>
                <CardFooter className="flex-col gap-2">
                    <Button 
                        type="submit" 
                        className="w-full"
                        onClick={handleSubmit}
                        disabled={isLoading}
                    >
                        {isLoading ? "Loading..." : "Sign In"}
                    </Button>
                </CardFooter>
            </Card>
        </div>
    )
}