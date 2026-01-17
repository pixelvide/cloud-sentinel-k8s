"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ShieldCheck } from "lucide-react";

export default function LoginPage() {
    const handleLogin = () => {
        window.location.href = "/api/v1/auth/login";
    };

    return (
        <div className="flex items-center justify-center min-h-screen bg-background relative overflow-hidden">
            {/* Background decorative elements */}
            <div className="absolute top-0 left-0 w-full h-full">
                <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary/5 rounded-full blur-[120px]" />
                <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-primary/5 rounded-full blur-[120px]" />
            </div>

            <Card className="w-full max-w-[450px] border-none shadow-2xl shadow-black/10 bg-card/60 backdrop-blur-2xl rounded-[2rem] md:rounded-[3rem] overflow-hidden relative z-10 mx-4">
                <CardHeader className="text-center p-8 md:p-12 pb-6">
                    <div className="flex justify-center mb-8">
                        <div className="p-4 bg-primary rounded-3xl shadow-2xl shadow-primary/40 rotate-12 hover:rotate-0 transition-transform duration-500">
                            <ShieldCheck className="w-10 h-10 text-primary-foreground text-white" />
                        </div>
                    </div>
                    <CardTitle className="text-3xl font-extrabold tracking-tight text-foreground">
                        Cloud Sentinel
                    </CardTitle>
                    <CardDescription className="text-xs font-bold uppercase tracking-[0.3em] opacity-40 mt-2">
                        K8s Access Console
                    </CardDescription>
                </CardHeader>
                <CardContent className="px-12 pb-12 pt-6">
                    <div className="space-y-6">
                        <p className="text-center text-sm text-muted-foreground leading-relaxed">
                            A secure, professional-grade k8s access environment for managing your Kubernetes clusters at
                            scale.
                        </p>
                        <Button
                            className="w-full h-14 font-extrabold text-sm uppercase tracking-widest shadow-xl shadow-primary/20 rounded-2xl transition-all active:scale-95 group"
                            onClick={handleLogin}
                        >
                            <span className="group-hover:translate-x-1 transition-transform inline-block mr-2">
                                Secure SSO Login
                            </span>
                            &rarr;
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
