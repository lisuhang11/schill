"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { LogIn, UserPlus } from "lucide-react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { loginUser, registerUser } from "@/lib/api";

const authSchema = z.object({
  username: z.string().trim().min(3, "用户名至少 3 个字符").max(32, "用户名不能超过 32 个字符"),
  password: z.string().min(6, "密码至少 6 个字符").max(64, "密码不能超过 64 个字符")
});

type AuthValues = z.infer<typeof authSchema>;

export function AuthForm({ mode }: { mode: "login" | "register" }) {
  const router = useRouter();
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();
  const {
    register,
    handleSubmit,
    formState: { errors }
  } = useForm<AuthValues>({
    resolver: zodResolver(authSchema),
    defaultValues: { username: "", password: "" }
  });

  function onSubmit(values: AuthValues) {
    setMessage("");
    startTransition(async () => {
      if (mode === "login") {
        const result = await loginUser(values);
        if (!result.ok) {
          setMessage(result.message);
          return;
        }
        window.localStorage.setItem("schill:userId", String(result.data.userId));
        window.localStorage.setItem("schill:accessToken", result.data.accessToken);
        window.localStorage.setItem("schill:refreshToken", result.data.refreshToken);
        router.push("/");
        return;
      }

      const result = await registerUser(values);
      if (!result.ok) {
        setMessage(result.message);
        return;
      }
      setMessage(`注册成功，用户 ID：${result.data.userId}。现在可以登录。`);
    });
  }

  const Icon = mode === "login" ? LogIn : UserPlus;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="mt-6 space-y-4">
      <div>
        <label className="text-sm font-medium text-marine-text" htmlFor="username">
          用户名
        </label>
        <input
          id="username"
          autoComplete="username"
          {...register("username")}
          className="focus-ring mt-2 h-11 w-full rounded-lg border border-[rgba(77,100,124,0.18)] px-3 text-sm"
        />
        {errors.username ? <p className="mt-1 text-xs text-red-600">{errors.username.message}</p> : null}
      </div>
      <div>
        <label className="text-sm font-medium text-marine-text" htmlFor="password">
          密码
        </label>
        <input
          id="password"
          type="password"
          autoComplete={mode === "login" ? "current-password" : "new-password"}
          {...register("password")}
          className="focus-ring mt-2 h-11 w-full rounded-lg border border-[rgba(77,100,124,0.18)] px-3 text-sm"
        />
        {errors.password ? <p className="mt-1 text-xs text-red-600">{errors.password.message}</p> : null}
      </div>
      <button
        type="submit"
        disabled={isPending}
        className="focus-ring inline-flex h-11 w-full items-center justify-center gap-2 rounded-lg bg-marine-deep px-4 text-sm font-semibold text-white disabled:opacity-60"
      >
        <Icon size={18} /> {mode === "login" ? "登录" : "注册"}
      </button>
      {message ? <p className="text-sm text-marine-deep">{message}</p> : null}
    </form>
  );
}
