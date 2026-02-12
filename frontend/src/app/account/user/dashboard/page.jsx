"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import ChatWidget from "@/components/utils/ChatWidget"
import { NavbarSignInOut } from "@/components/utils/NavbarSignInOut"
import axios from "@/lib/axios"

export default function UserPage() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    const user = localStorage.getItem('user')
    
    if (!token || !user) {
      router.push('/auth/login')
      return
    }

    axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
  }, [router])

  return (
    <>
      <NavbarSignInOut />
      <main className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold">Halaman User</h1>
        <p className="mt-4 text-muted-foreground">
          Selamat datang di dashboard user
        </p>
      </main>
      <ChatWidget />
    </>
  )
}