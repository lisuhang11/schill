import Link from "next/link";
import { AuthForm } from "@/components/AuthForm";

export default function RegisterPage() {
  return (
    <main className="mx-auto w-full max-w-xl px-4 py-10 md:px-8">
      <section className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-6 shadow-soft">
        <h1 className="text-3xl font-semibold text-marine-text">注册</h1>
        <p className="mt-2 text-sm leading-6 text-marine-muted">
          使用 gateway 的 `POST /api/auth/register` 创建账号，注册成功后可返回登录。
        </p>
        <AuthForm mode="register" />
        <p className="mt-5 text-center text-sm text-marine-muted">
          已有账号？{" "}
          <Link className="font-semibold text-marine-deep" href="/login">
            去登录
          </Link>
        </p>
      </section>
    </main>
  );
}
