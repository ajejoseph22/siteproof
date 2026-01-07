import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { motion } from 'framer-motion';
import { HardHat, Mail, Lock, Eye, EyeOff, Sun, Moon, ArrowRight, User, Building2, Phone } from 'lucide-react';
import { Button, Input } from '@/components/ui';
import { useAppStore } from '@/stores/app';
import { cn } from '@/lib/utils';

const registerSchema = z.object({
  companyName: z.string().min(2, 'Company name must be at least 2 characters'),
  name: z.string().min(2, 'Name must be at least 2 characters'),
  email: z.string().email('Please enter a valid email'),
  phone: z.string().optional(),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ['confirmPassword'],
});

type RegisterFormData = z.infer<typeof registerSchema>;

export function RegisterPage() {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const { register: registerUser, isDarkMode, toggleDarkMode } = useAppStore();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      companyName: '',
      name: '',
      email: '',
      phone: '',
      password: '',
      confirmPassword: '',
    },
  });

  const onSubmit = async (data: RegisterFormData) => {
    setIsLoading(true);
    setError('');

    try {
      const success = await registerUser({
        companyName: data.companyName,
        name: data.name,
        email: data.email,
        phone: data.phone || undefined,
        password: data.password,
      });
      if (success) {
        navigate('/dashboard');
      } else {
        setError('Registration failed. Please try again.');
      }
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('An error occurred. Please try again.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex flex-col bg-charcoal-50 dark:bg-charcoal-950">
      {/* Background Pattern */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-1/4 -right-1/4 w-1/2 h-1/2 bg-gradient-to-br from-safety-500/5 to-transparent rounded-full blur-3xl" />
        <div className="absolute -bottom-1/4 -left-1/4 w-1/2 h-1/2 bg-gradient-to-tr from-safety-500/5 to-transparent rounded-full blur-3xl" />
      </div>

      {/* Header */}
      <header className="relative flex items-center justify-between p-4 sm:p-6">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-safety-500 to-safety-600 flex items-center justify-center shadow-lg shadow-safety-500/25">
            <HardHat className="w-6 h-6 text-white" />
          </div>
          <span className="font-display text-2xl tracking-wide text-charcoal-900 dark:text-white">
            SITEPROOF
          </span>
        </div>
        <button
          onClick={toggleDarkMode}
          className="p-2.5 rounded-lg bg-white dark:bg-charcoal-800 border-2 border-charcoal-200 dark:border-charcoal-700 hover:border-safety-500 transition-colors shadow-sm"
          aria-label="Toggle theme"
        >
          {isDarkMode ? (
            <Sun className="w-5 h-5 text-safety-500" />
          ) : (
            <Moon className="w-5 h-5 text-charcoal-600" />
          )}
        </button>
      </header>

      {/* Main Content */}
      <main className="relative flex-1 flex items-center justify-center p-4 py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: 'easeOut' }}
          className="w-full max-w-md"
        >
          {/* Construction Tape Decoration */}
          <div className="h-2 rounded-t-2xl bg-gradient-to-r from-safety-500 via-charcoal-900 to-safety-500 dark:via-charcoal-950" />

          {/* Register Card */}
          <div className="bg-white dark:bg-charcoal-900 rounded-b-2xl p-6 sm:p-8 shadow-2xl shadow-charcoal-900/10 dark:shadow-charcoal-950/50 border-2 border-t-0 border-charcoal-200 dark:border-charcoal-800">
            {/* Header */}
            <div className="text-center mb-6">
              <motion.div
                initial={{ scale: 0.9, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ delay: 0.2, duration: 0.3 }}
                className="w-16 h-16 rounded-2xl bg-gradient-to-br from-safety-100 to-safety-50 dark:from-safety-900/30 dark:to-safety-950/30 flex items-center justify-center mx-auto mb-4 border-2 border-safety-200 dark:border-safety-800/50"
              >
                <HardHat className="w-8 h-8 text-safety-500" />
              </motion.div>
              <h1 className="text-2xl sm:text-3xl font-display tracking-wide text-charcoal-900 dark:text-white mb-2">
                CREATE ACCOUNT
              </h1>
              <p className="text-charcoal-500 dark:text-charcoal-400 text-sm">
                Get started with your T&M ticket management
              </p>
            </div>

            {/* Error Message */}
            {error && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                className="mb-6 p-4 rounded-xl bg-red-50 dark:bg-red-900/20 border-2 border-red-200 dark:border-red-800/50 text-red-700 dark:text-red-400 text-sm font-medium flex items-center gap-2"
              >
                <svg className="w-5 h-5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
                {error}
              </motion.div>
            )}

            {/* Register Form */}
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <Input
                label="Company Name"
                type="text"
                placeholder="Apex Electric Co."
                leftIcon={<Building2 className="w-5 h-5" />}
                error={errors.companyName?.message}
                {...register('companyName')}
              />

              <Input
                label="Your Name"
                type="text"
                placeholder="John Smith"
                leftIcon={<User className="w-5 h-5" />}
                error={errors.name?.message}
                {...register('name')}
              />

              <Input
                label="Email Address"
                type="email"
                placeholder="you@company.com"
                leftIcon={<Mail className="w-5 h-5" />}
                error={errors.email?.message}
                {...register('email')}
              />

              <Input
                label="Phone (optional)"
                type="tel"
                placeholder="+1 (555) 123-4567"
                leftIcon={<Phone className="w-5 h-5" />}
                error={errors.phone?.message}
                {...register('phone')}
              />

              <Input
                label="Password"
                type={showPassword ? 'text' : 'password'}
                placeholder="Minimum 8 characters"
                leftIcon={<Lock className="w-5 h-5" />}
                rightIcon={
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="hover:text-safety-500 transition-colors"
                  >
                    {showPassword ? (
                      <EyeOff className="w-5 h-5" />
                    ) : (
                      <Eye className="w-5 h-5" />
                    )}
                  </button>
                }
                error={errors.password?.message}
                {...register('password')}
              />

              <Input
                label="Confirm Password"
                type={showConfirmPassword ? 'text' : 'password'}
                placeholder="Repeat your password"
                leftIcon={<Lock className="w-5 h-5" />}
                rightIcon={
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    className="hover:text-safety-500 transition-colors"
                  >
                    {showConfirmPassword ? (
                      <EyeOff className="w-5 h-5" />
                    ) : (
                      <Eye className="w-5 h-5" />
                    )}
                  </button>
                }
                error={errors.confirmPassword?.message}
                {...register('confirmPassword')}
              />

              <div className="flex items-start gap-2.5 text-sm">
                <div className="relative mt-0.5">
                  <input
                    type="checkbox"
                    id="terms"
                    className="peer sr-only"
                    required
                  />
                  <div className={cn(
                    "w-5 h-5 rounded-md border-2 border-charcoal-300 dark:border-charcoal-600",
                    "bg-white dark:bg-charcoal-800",
                    "peer-checked:bg-safety-500 peer-checked:border-safety-500",
                    "transition-all cursor-pointer"
                  )} />
                  <svg
                    className="absolute top-0.5 left-0.5 w-4 h-4 text-white opacity-0 peer-checked:opacity-100 transition-opacity pointer-events-none"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    strokeWidth={3}
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <label htmlFor="terms" className="text-charcoal-600 dark:text-charcoal-400 cursor-pointer">
                  I agree to the{' '}
                  <button type="button" className="text-safety-500 hover:text-safety-600 dark:hover:text-safety-400 font-medium hover:underline">
                    Terms of Service
                  </button>{' '}
                  and{' '}
                  <button type="button" className="text-safety-500 hover:text-safety-600 dark:hover:text-safety-400 font-medium hover:underline">
                    Privacy Policy
                  </button>
                </label>
              </div>

              <Button
                type="submit"
                className="w-full mt-6"
                size="xl"
                isLoading={isLoading}
                rightIcon={!isLoading && <ArrowRight className="w-5 h-5" />}
              >
                Create Account
              </Button>
            </form>

            {/* Sign In Link */}
            <div className="mt-6 text-center">
              <p className="text-sm text-charcoal-500 dark:text-charcoal-400">
                Already have an account?{' '}
                <Link
                  to="/login"
                  className="text-safety-500 hover:text-safety-600 dark:hover:text-safety-400 font-medium hover:underline transition-colors"
                >
                  Sign in
                </Link>
              </p>
            </div>
          </div>

          {/* Decorative bottom tape */}
          <div className="flex gap-1 mt-4 justify-center">
            {[...Array(5)].map((_, i) => (
              <div
                key={i}
                className="w-8 h-1 rounded-full bg-gradient-to-r from-safety-500 to-safety-600"
                style={{ opacity: 1 - i * 0.15 }}
              />
            ))}
          </div>
        </motion.div>
      </main>

      {/* Footer */}
      <footer className="relative text-center p-4 text-sm text-charcoal-500 dark:text-charcoal-400">
        <p>&copy; {new Date().getFullYear()} SiteProof. Built for construction professionals.</p>
      </footer>
    </div>
  );
}
