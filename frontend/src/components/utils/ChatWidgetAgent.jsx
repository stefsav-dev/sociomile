"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import axios from "@/lib/axios";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "../ui/card";
import { Input } from "../ui/input";
import { MessageCircle, X, Send, Bell } from "lucide-react";
import { Button } from "../ui/button";
import { motion, AnimatePresence } from "framer-motion";

export default function ChatWidgetAgent() {
    const router = useRouter();
    const [isOpen, setIsOpen] = useState(false);
    const [channels, setChannels] = useState([]);
    const [selectedChannel, setSelectedChannel] = useState(null);
    const [messages, setMessages] = useState([]);
    const [inputMessage, setInputMessage] = useState("");
    const [isSending, setIsSending] = useState(false);
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [isAuthChecking, setIsAuthChecking] = useState(true);
    const [unreadCount, setUnreadCount] = useState(0);
    const [channelUnread, setChannelUnread] = useState({});
    const messagesEndRef = useRef(null);
    const pollingInterval = useRef(null);
    const channelsPollingInterval = useRef(null);
    const audioRef = useRef(null);

    useEffect(() => {
        const token = localStorage.getItem('token');
        const user = localStorage.getItem('user');
        
        if (token && user) {
            setIsAuthenticated(true);
            axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
            
            const userData = JSON.parse(user);
            if (userData.role !== 'agent') {
                console.log('Bukan agent, tidak bisa akses chat agent');
                setIsAuthenticated(false);
            }
        } else {
            setIsAuthenticated(false);
        }
        setIsAuthChecking(false);
    }, []);


    useEffect(() => {
        if (typeof window !== 'undefined' && 'Notification' in window) {
            if (Notification.permission !== 'granted' && Notification.permission !== 'denied') {
                Notification.requestPermission();
            }
        }
    }, []);

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    useEffect(() => {
        if (isAuthenticated && isOpen) {
            fetchAllChannels();
            
            channelsPollingInterval.current = setInterval(() => {
                fetchAllChannels();
            }, 5000);
        }
        
        return () => {
            if (channelsPollingInterval.current) {
                clearInterval(channelsPollingInterval.current);
            }
            if (pollingInterval.current) {
                clearInterval(pollingInterval.current);
            }
        };
    }, [isOpen, isAuthenticated]);

    useEffect(() => {
        if (selectedChannel?.id && selectedChannel.status !== "closed" && isAuthenticated) {
            fetchMessages(selectedChannel.id);
            
            pollingInterval.current = setInterval(() => {
                fetchMessages(selectedChannel.id);
            }, 3000);
        }

        return () => {
            if (pollingInterval.current) {
                clearInterval(pollingInterval.current);
                pollingInterval.current = null;
            }
        };
    }, [selectedChannel?.id, selectedChannel?.status, isAuthenticated]);

    const fetchAllChannels = async () => {
        try {
            const availableRes = await axios.get('/agent/channels/available');
            const availableChannels = availableRes.data.data || [];
            
            const assignedRes = await axios.get('/agent/conversations?status=open,assigned');
            const assignedChannels = assignedRes.data.data || [];
            

            const allChannels = [...availableChannels, ...assignedChannels];
            

            for (const ch of allChannels) {
                await checkChannelNewMessages(ch.id);
            }
            
            setChannels(allChannels);


            if (assignedChannels.length > 0 && !selectedChannel) {
                setSelectedChannel(assignedChannels[0]);
                fetchMessages(assignedChannels[0].id);
            }
            

            const totalUnread = Object.values(channelUnread).reduce((a, b) => a + b, 0);
            setUnreadCount(totalUnread);
            
        } catch (error) {
            console.error('Failed to fetch channels:', error);
        }
    };

    const checkChannelNewMessages = async (channelId) => {
        try {
            const response = await axios.get(`/agent/channels/${channelId}`);
            const newMessages = response.data.data.messages || [];
            
            const oldMessages = messages;
            if (selectedChannel?.id !== channelId) {
                const lastMsg = newMessages[newMessages.length - 1];
                if (lastMsg && lastMsg.sender_type === 'customer' && !lastMsg.is_read) {
                    setChannelUnread(prev => ({
                        ...prev,
                        [channelId]: (prev[channelId] || 0) + 1
                    }));
                    
                    if (Notification.permission === 'granted') {
                        new Notification('Pesan Baru dari Customer', {
                            body: lastMsg.message,
                            icon: '/logo.png'
                        });
                    }
                    

                    if (audioRef.current) {
                        audioRef.current.play().catch(e => console.log('Audio play failed:', e));
                    }
                }
            }
        } catch (error) {
            console.error('Failed to check channel messages:', error);
        }
    };

    const fetchMessages = async (channelId) => {
        try {
            const response = await axios.get(`/agent/channels/${channelId}`);
            const newMessages = response.data.data.messages || [];
            

            setChannelUnread(prev => ({
                ...prev,
                [channelId]: 0
            }));
            
            setMessages((prevMessages) => {
                if (JSON.stringify(prevMessages) !== JSON.stringify(newMessages)) {
                    axios.put(`/agent/channels/${channelId}/read`).catch(console.error);
                    return newMessages;
                }
                return prevMessages;
            });
        } catch (error) {
            console.error('Failed to fetch messages:', error);
        }
    };

    const assignChannel = async (channelId) => {
        try {
            const response = await axios.patch(`/agent/channels/${channelId}/assign`);
            if (response.data.success) {
                fetchAllChannels();
                
                const newChannel = response.data.data;
                setSelectedChannel({
                    id: newChannel.channel_id,
                    assigned_agent_id: newChannel.assigned_agent_id,
                    status: newChannel.status
                });
                
                setChannelUnread(prev => ({
                    ...prev,
                    [channelId]: 0
                }));
            }
        } catch (error) {
            console.error('Failed to assign channel:', error);
        }
    };

    const sendMessage = async () => {
        if (!inputMessage.trim() || !selectedChannel?.id || !isAuthenticated) return;

        try {
            setIsSending(true);
            const response = await axios.post(`/agent/channels/${selectedChannel.id}/messages`, {
                message: inputMessage.trim()
            });

            setMessages((prev) => [...prev, response.data.data]);
            setInputMessage("");
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
        
        sendMessage();
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

    const getInitials = (name) => {
        return name?.charAt(0).toUpperCase() || "C";
    };

    useEffect(() => {
        const total = Object.values(channelUnread).reduce((a, b) => a + b, 0);
        setUnreadCount(total);
    }, [channelUnread]);

    if (isAuthChecking) {
        return null;
    }

    if (!isOpen) {
        return (
            <motion.div
                whileTap={{ scale: 0.9 }}
                whileHover={{ scale: 1.1 }}
                transition={{ type: "spring", stiffness: 400, damping: 17 }}
                className="fixed bottom-4 right-24 z-50"
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
                    <Bell className="h-6 w-6" />
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
        <div className="fixed bottom-4 right-24 z-50 flex flex-col items-end">
            <AnimatePresence>
                <motion.div
                    initial={{ opacity: 0, y: 20, scale: 0.95 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    exit={{ opacity: 0, y: 20, scale: 0.95 }}
                    transition={{ duration: 0.2, ease: "easeOut" }}
                    className="mb-4"
                >
                    <Card className="w-[800px] h-[600px] shadow-lg flex overflow-hidden rounded-lg">
                        <div className="w-64 border-r bg-muted/10">
                            <CardHeader className="px-4 py-3 border-b">
                                <CardTitle className="text-sm font-medium">
                                    Daftar Percakapan
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="p-2 overflow-y-auto h-[calc(600px-60px)]">
                                {channels.length === 0 ? (
                                    <div className="text-center py-8">
                                        <p className="text-sm text-muted-foreground">
                                            Tidak ada percakapan
                                        </p>
                                    </div>
                                ) : (
                                    <div className="space-y-2">
                                        {channels.map((ch) => (
                                            <button
                                                key={ch.id}
                                                onClick={() => {
                                                    if (ch.assigned_agent_id === 0) {
                                                        assignChannel(ch.id);
                                                    } else {
                                                        setSelectedChannel(ch);
                                                        fetchMessages(ch.id);
                                                        setChannelUnread(prev => ({
                                                            ...prev,
                                                            [ch.id]: 0
                                                        }));
                                                    }
                                                }}
                                                className={`w-full p-3 rounded-lg text-left transition-colors relative ${
                                                    selectedChannel?.id === ch.id
                                                        ? 'bg-primary text-primary-foreground'
                                                        : 'hover:bg-muted'
                                                }`}
                                            >
                                                <div className="flex items-center gap-2">
                                                    <div className={`h-8 w-8 rounded-full flex items-center justify-center ${
                                                        selectedChannel?.id === ch.id
                                                            ? 'bg-primary-foreground/20'
                                                            : 'bg-muted-foreground/20'
                                                    }`}>
                                                        <span className="text-xs font-semibold">
                                                            {ch.customer_name?.charAt(0) || "C"}
                                                        </span>
                                                    </div>
                                                    <div className="flex-1 min-w-0">
                                                        <p className="text-sm font-medium truncate">
                                                            {ch.customer_name || "Customer"}
                                                        </p>
                                                        <p className="text-xs opacity-70 truncate">
                                                            {ch.assigned_agent_id === 0 ? (
                                                                <span className="text-yellow-600">Menunggu</span>
                                                            ) : (
                                                                <span className="text-green-600">Terhubung</span>
                                                            )}
                                                        </p>
                                                    </div>
                                                    {channelUnread[ch.id] > 0 && (
                                                        <span className="h-5 w-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
                                                            {channelUnread[ch.id]}
                                                        </span>
                                                    )}
                                                </div>
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </CardContent>
                        </div>

                        <div className="flex-1 flex flex-col">

                            <div className="bg-primary text-primary-foreground">
                                <CardHeader className="flex flex-row items-center justify-between px-4 py-3">
                                    <div className="flex items-center gap-2">
                                        <div className="h-8 w-8 rounded-full bg-primary-foreground/20 flex items-center justify-center">
                                            <span className="text-sm font-semibold">
                                                {selectedChannel?.customer_name?.charAt(0) || "C"}
                                            </span>
                                        </div>
                                        <div>
                                            <CardTitle className="text-sm font-medium">
                                                {selectedChannel?.customer_name || "Customer Service"}
                                            </CardTitle>
                                            <p className="text-xs text-primary-foreground/80">
                                                {!selectedChannel ? (
                                                    "Pilih percakapan"
                                                ) : selectedChannel.assigned_agent_id === 0 ? (
                                                    "Menunggu di-assign"
                                                ) : (
                                                    "Online"
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
                                {!selectedChannel ? (
                                    <div className="flex flex-col items-center justify-center h-full text-center">
                                        <MessageCircle className="h-12 w-12 text-muted-foreground mb-4" />
                                        <h4 className="font-semibold mb-2">Pilih Percakapan</h4>
                                        <p className="text-sm text-muted-foreground">
                                            Pilih percakapan dari daftar di samping
                                        </p>
                                    </div>
                                ) : messages.length === 0 ? (
                                    <div className="flex flex-col items-center justify-center h-full text-center">
                                        <p className="text-sm text-muted-foreground">
                                            Belum ada pesan. Mulai percakapan!
                                        </p>
                                    </div>
                                ) : (
                                    messages.map((message) => (
                                        <div
                                            key={message.id}
                                            className={`flex ${message.sender_type === 'agent' ? 'justify-end' : 'justify-start'}`}
                                        >
                                            <div
                                                className={`max-w-[80%] ${
                                                    message.sender_type === 'agent'
                                                        ? 'bg-primary text-primary-foreground rounded-lg rounded-tr-none'
                                                        : 'bg-muted rounded-lg rounded-tl-none'
                                                } p-3 ${
                                                    message.sender_type === 'customer' && !message.is_read
                                                        ? 'ring-2 ring-blue-500'
                                                        : ''
                                                }`}
                                            >
                                                <p className="text-sm break-words">{message.message}</p>
                                                <span className={`text-[10px] opacity-70 block mt-1 text-right ${
                                                    message.sender_type === 'agent'
                                                        ? 'text-primary-foreground/70'
                                                        : 'text-muted-foreground'
                                                }`}>
                                                    {formatTime(message.created_at)}
                                                    {message.sender_type === 'customer' && !message.is_read && (
                                                        <span className="ml-2 text-blue-500">● Baru</span>
                                                    )}
                                                </span>
                                            </div>
                                        </div>
                                    ))
                                )}
                                <div ref={messagesEndRef} />
                            </CardContent>
                            
                            {selectedChannel && selectedChannel.assigned_agent_id === 0 ? (
                                <CardFooter className="p-2 border-t">
                                    <Button 
                                        onClick={() => assignChannel(selectedChannel.id)}
                                        className="w-full"
                                        variant="default"
                                    >
                                        Ambil Percakapan
                                    </Button>
                                </CardFooter>
                            ) : selectedChannel && selectedChannel.status !== "closed" && isAuthenticated && (
                                <CardFooter className="p-2 border-t">
                                    <form onSubmit={handleSendMessage} className="flex w-full gap-2">
                                        <Input 
                                            placeholder="Tulis Pesan..." 
                                            className="text-sm flex-grow"
                                            value={inputMessage}
                                            onChange={(e) => setInputMessage(e.target.value)}
                                            onKeyPress={handleKeyPress}
                                            disabled={isSending}
                                        />
                                        <Button 
                                            type="submit"
                                            size="icon" 
                                            className="h-10 w-10"
                                            disabled={!inputMessage.trim() || isSending}
                                        >
                                            <Send className="h-4 w-4" />
                                        </Button>
                                    </form>
                                </CardFooter>
                            )}
                        </div>
                    </Card>
                </motion.div>
            </AnimatePresence>
        </div>
    );
}