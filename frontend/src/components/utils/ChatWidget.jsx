"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import axios from "@/lib/axios";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "../ui/card";
import { Input } from "../ui/input";
import { MessageCircle, X, Send, Bell } from "lucide-react";
import { Button } from "../ui/button";
import { motion, AnimatePresence } from "framer-motion";

export default function ChatWidget() {
    const router = useRouter();
    const [isOpen, setIsOpen] = useState(false);
    const [channel, setChannel] = useState(null);
    const [messages, setMessages] = useState([]);
    const [inputMessage, setInputMessage] = useState("");
    const [isSending, setIsSending] = useState(false);
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [isAuthChecking, setIsAuthChecking] = useState(true);
    const [unreadCount, setUnreadCount] = useState(0);
    const [lastMessageId, setLastMessageId] = useState(null);
    const messagesEndRef = useRef(null);
    const pollingInterval = useRef(null);
    const audioRef = useRef(null);


    useEffect(() => {
        const token = localStorage.getItem('token');
        const user = localStorage.getItem('user');
        
        if (token && user) {
            setIsAuthenticated(true);
            axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
        } else {
            setIsAuthenticated(false);
        }
        setIsAuthChecking(false);
    }, []);


    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);


    useEffect(() => {
        if (isAuthenticated && !channel) {
            checkExistingChannel();
        }
    }, [isAuthenticated]);


    useEffect(() => {
        if (channel?.id && channel.status !== "closed" && isAuthenticated) {
            checkNewMessages();
            
            pollingInterval.current = setInterval(() => {
                checkNewMessages();
            }, 5000);
        }

        return () => {
            if (pollingInterval.current) {
                clearInterval(pollingInterval.current);
                pollingInterval.current = null;
            }
        };
    }, [channel?.id, channel?.status, isAuthenticated]);

    const checkExistingChannel = async () => {
        try {

            const response = await axios.get('/user/channels?status=open,assigned');
            const channels = response.data.data || [];
            
            if (channels.length > 0) {
                const existingChannel = channels[0];
                setChannel(existingChannel);
                

                const msgResponse = await axios.get(`/user/channels/${existingChannel.id}`);
                setMessages(msgResponse.data.data.messages || []);
                
   
                const msgs = msgResponse.data.data.messages || [];
                if (msgs.length > 0) {
                    setLastMessageId(msgs[msgs.length - 1].id);
                }
            }
        } catch (error) {
            console.error('Failed to check existing channel:', error);
        }
    };

    const checkNewMessages = async () => {
        if (!channel?.id) return;

        try {
            const response = await axios.get(`/user/channels/${channel.id}`);
            const newMessages = response.data.data.messages || [];
            

            if (response.data.data.channel) {
                setChannel(response.data.data.channel);
            }
            

            const oldMessages = messages;
            const lastMsg = oldMessages[oldMessages.length - 1];
            
            if (newMessages.length > oldMessages.length) {

                const newMsgs = newMessages.slice(oldMessages.length);
                

                const agentMessages = newMsgs.filter(msg => msg.sender_type === 'agent');
                
                if (agentMessages.length > 0) {

                    setUnreadCount(prev => prev + agentMessages.length);
                    
                    if (audioRef.current) {
                        audioRef.current.play().catch(e => console.log('Audio play failed:', e));
                    }
                    
                    if (Notification.permission === 'granted') {
                        new Notification('Pesan Baru dari Agent', {
                            body: agentMessages[0].message,
                            icon: '/logo.png'
                        });
                    }
                }
                
                setMessages(newMessages);
                setLastMessageId(newMessages[newMessages.length - 1].id);
            }
        } catch (error) {
            console.error('Failed to check new messages:', error);
        }
    };

    const createNewChannel = async () => {
        if (!inputMessage.trim() || !isAuthenticated) return;

        try {
            setIsSending(true);
            const response = await axios.post('/user/channels', {
                tenant_id: 1,
                message: inputMessage.trim()
            });

            const { channel_id, message_id } = response.data.data;
            
            setChannel({
                id: channel_id,
                status: 'open',
                assigned_agent_id: 0,
                customer_id: JSON.parse(localStorage.getItem('user')).id
            });
            
            const newMessage = {
                id: message_id,
                conversation_id: channel_id,
                sender_type: "customer",
                message: inputMessage.trim(),
                created_at: new Date().toISOString(),
                is_read: false
            };
            
            setMessages([newMessage]);
            setLastMessageId(message_id);
            setInputMessage("");
            setUnreadCount(0);
            
        } catch (error) {
            console.error('Failed to create channel:', error);
        } finally {
            setIsSending(false);
        }
    };

    const sendMessage = async () => {
        if (!inputMessage.trim() || !channel?.id || !isAuthenticated) return;

        try {
            setIsSending(true);
            const response = await axios.post(`/user/channels/${channel.id}/messages`, {
                message: inputMessage.trim()
            });

            setMessages((prev) => [...prev, response.data.data]);
            setLastMessageId(response.data.data.id);
            setInputMessage("");
            
            setUnreadCount(0);
            
        } catch (error) {
            console.error('Failed to send message:', error);
        } finally {
            setIsSending(false);
        }
    };

    const handleSendMessage = (e) => {
        e.preventDefault();
        
        if (!isAuthenticated) {
            router.push('/auth/login');
            return;
        }
        
        if (channel) {
            sendMessage();
        } else {
            createNewChannel();
        }
    };

    const handleKeyPress = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSendMessage(e);
        }
    };

    const formatTime = (timestamp) => {
        const date = new Date(timestamp);
        return date.toLocaleTimeString('id-ID', {
            hour: '2-digit',
            minute: '2-digit',
            hour12: false
        });
    };

    useEffect(() => {
        if (typeof window !== 'undefined' && 'Notification' in window) {
            if (Notification.permission !== 'granted' && Notification.permission !== 'denied') {
                Notification.requestPermission();
            }
        }
    }, []);

    useEffect(() => {
        if (isOpen) {
            setUnreadCount(0);
        }
    }, [isOpen]);

    if (isAuthChecking) {
        return null;
    }

    if (!isOpen) {
        return (
            <motion.div
                whileTap={{ scale: 0.9 }}
                whileHover={{ scale: 1.1 }}
                transition={{ type: "spring", stiffness: 400, damping: 17 }}
                className="fixed bottom-4 right-4 z-50"
            >
                <Button
                    onClick={() => {
                        if (!isAuthenticated) {
                            router.push('/auth/login');
                        } else {
                            setIsOpen(true);
                        }
                    }}
                    className="h-12 w-12 rounded-full shadow-lg relative overflow-hidden"
                    size="icon"
                >
                    <MessageCircle className="h-6 w-6" />
                    {unreadCount > 0 && (
                        <span className="absolute -top-1 -right-1 h-5 w-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
                            {unreadCount > 9 ? '9+' : unreadCount}
                        </span>
                    )}
                    <motion.div
                        className="absolute inset-0 rounded-full bg-primary/30"
                        initial={{ scale: 0, opacity: 0 }}
                        animate={{
                            scale: [0, 1.5, 1.5, 0],
                            opacity: [0, 0.3, 0.3, 0],
                        }}
                        transition={{
                            duration: 2,
                            repeat: Infinity,
                            repeatDelay: 1,
                            ease: "easeOut"
                        }}
                    />
                </Button>
            </motion.div>
        );
    }

    return (
        <div className="fixed bottom-4 right-4 z-50 flex flex-col items-end">
            <AnimatePresence>
                <motion.div
                    initial={{ opacity: 0, y: 20, scale: 0.95 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    exit={{ opacity: 0, y: 20, scale: 0.95 }}
                    transition={{ duration: 0.2, ease: "easeOut" }}
                    className="mb-4"
                >
                    <Card className="w-96 h-[600px] shadow-lg flex flex-col overflow-hidden rounded-lg">

                        <div className="bg-primary text-primary-foreground">
                            <CardHeader className="flex flex-row items-center justify-between px-4 py-3">
                                <div className="flex items-center gap-2">
                                    <div className="h-8 w-8 rounded-full bg-primary-foreground/20 flex items-center justify-center">
                                        <span className="text-sm font-semibold">CS</span>
                                    </div>
                                    <div>
                                        <CardTitle className="text-sm font-medium">
                                            Chat Support
                                        </CardTitle>
                                        <p className="text-xs text-primary-foreground/80">
                                            {!isAuthenticated ? (
                                                "Silakan login"
                                            ) : !channel ? (
                                                "Mulai percakapan"
                                            ) : channel.status === "closed" ? (
                                                "Percakapan selesai"
                                            ) : channel.assigned_agent_id ? (
                                                "Terhubung dengan agent"
                                            ) : (
                                                "Menunggu agent..."
                                            )}
                                        </p>
                                    </div>
                                </div>
                                <Button 
                                    size="icon" 
                                    variant="ghost" 
                                    className="h-7 w-7 hover:bg-primary-foreground/20 rounded-full" 
                                    onClick={() => setIsOpen(false)}
                                >
                                    <X className="h-4 w-4"/>
                                </Button>
                            </CardHeader>
                        </div>
                        
                        <CardContent className="flex-grow p-4 overflow-y-auto space-y-4">
                            {!isAuthenticated ? (
                                <div className="flex flex-col items-center justify-center h-full text-center">
                                    <MessageCircle className="h-12 w-12 text-muted-foreground mb-4" />
                                    <h4 className="font-semibold mb-2">Silakan Login</h4>
                                    <p className="text-sm text-muted-foreground mb-4">
                                        Anda perlu login untuk menggunakan fitur chat
                                    </p>
                                    <Button onClick={() => router.push('/auth/login')}>
                                        Login Sekarang
                                    </Button>
                                </div>
                            ) : messages.length === 0 ? (
                                <div className="flex flex-col items-center justify-center h-full text-center">
                                    <MessageCircle className="h-12 w-12 text-muted-foreground mb-4" />
                                    <h4 className="font-semibold mb-2">Hallo! ada yang bisa dibantu?</h4>
                                    <p className="text-sm text-muted-foreground">
                                        Kirim pesan untuk memulai percakapan dengan agent kami
                                    </p>
                                </div>
                            ) : (
                                messages.map((message) => (
                                    <div
                                        key={message.id}
                                        className={`flex ${message.sender_type === 'customer' ? 'justify-end' : 'justify-start'}`}
                                    >
                                        <div
                                            className={`max-w-[80%] ${
                                                message.sender_type === 'customer'
                                                    ? 'bg-primary text-primary-foreground rounded-lg rounded-tr-none'
                                                    : 'bg-muted rounded-lg rounded-tl-none'
                                            } p-3 ${
                                                message.sender_type === 'agent' && 
                                                message.id > (lastMessageId - messages.length) && 
                                                !message.is_read 
                                                    ? 'ring-2 ring-blue-500' 
                                                    : ''
                                            }`}
                                        >
                                            <p className="text-sm break-words">{message.message}</p>
                                            <span className={`text-[10px] opacity-70 block mt-1 text-right ${
                                                message.sender_type === 'customer'
                                                    ? 'text-primary-foreground/70'
                                                    : 'text-muted-foreground'
                                            }`}>
                                                {formatTime(message.created_at)}
                                                {message.sender_type === 'agent' && !message.is_read && (
                                                    <span className="ml-2 text-blue-500">● Baru</span>
                                                )}
                                            </span>
                                        </div>
                                    </div>
                                ))
                            )}
                            <div ref={messagesEndRef} />
                        </CardContent>
                        
                        <CardFooter className="p-2 border-t">
                            <form onSubmit={handleSendMessage} className="flex w-full gap-2">
                                <Input 
                                    placeholder={isAuthenticated ? "Tulis Pesan..." : "Login untuk chat"}
                                    className="text-sm flex-grow"
                                    value={inputMessage}
                                    onChange={(e) => setInputMessage(e.target.value)}
                                    onKeyPress={handleKeyPress}
                                    disabled={!isAuthenticated || isSending || channel?.status === "closed"}
                                />
                                <Button 
                                    type="submit"
                                    size="icon" 
                                    className="h-10 w-10"
                                    disabled={!isAuthenticated || !inputMessage.trim() || isSending}
                                >
                                    <Send className="h-4 w-4" />
                                </Button>
                            </form>
                        </CardFooter>
                    </Card>
                </motion.div>
            </AnimatePresence>
        </div>
    );
}