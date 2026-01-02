import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import {
  ClipboardList,
  Clock,
  CheckCircle2,
  FileText,
  DollarSign,
  TrendingUp,
  ArrowRight,
  CalendarDays,
  AlertCircle,
  Zap,
} from 'lucide-react';
import { Card, CardHeader, CardTitle, StatusBadge, Button } from '@/components/ui';
import { useAppStore } from '@/stores/app';
import { formatCurrency, formatRelativeTime, cn } from '@/lib/utils';
import { mockDashboardStats } from '@/data/mock';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.08 },
  },
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease: [0.25, 0.46, 0.45, 0.94] as const }
  },
};

export function DashboardPage() {
  const { user, tickets } = useAppStore();
  const stats = mockDashboardStats;

  const recentTickets = tickets.slice(0, 5);

  const statCards = [
    {
      title: 'Total Tickets',
      value: stats.totalTickets,
      icon: ClipboardList,
      color: 'text-blue-500',
      bgColor: 'bg-blue-500/10 dark:bg-blue-500/20',
      borderColor: 'border-blue-200 dark:border-blue-800/50',
    },
    {
      title: 'Pending Review',
      value: stats.pendingTickets,
      icon: Clock,
      color: 'text-amber-500',
      bgColor: 'bg-amber-500/10 dark:bg-amber-500/20',
      borderColor: 'border-amber-200 dark:border-amber-800/50',
    },
    {
      title: 'Approved',
      value: stats.approvedTickets,
      icon: CheckCircle2,
      color: 'text-emerald-500',
      bgColor: 'bg-emerald-500/10 dark:bg-emerald-500/20',
      borderColor: 'border-emerald-200 dark:border-emerald-800/50',
    },
    {
      title: 'Draft',
      value: stats.draftTickets,
      icon: FileText,
      color: 'text-charcoal-500 dark:text-charcoal-400',
      bgColor: 'bg-charcoal-500/10 dark:bg-charcoal-500/20',
      borderColor: 'border-charcoal-200 dark:border-charcoal-700',
    },
  ];

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'ticket_approved':
        return <CheckCircle2 className="w-4 h-4 text-emerald-500" />;
      case 'ticket_rejected':
        return <AlertCircle className="w-4 h-4 text-red-500" />;
      case 'ticket_submitted':
        return <Clock className="w-4 h-4 text-amber-500" />;
      default:
        return <FileText className="w-4 h-4 text-charcoal-400" />;
    }
  };

  const greeting = () => {
    const hour = new Date().getHours();
    if (hour < 12) return 'Good Morning';
    if (hour < 18) return 'Good Afternoon';
    return 'Good Evening';
  };

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className="space-y-6"
    >
      {/* Header */}
      <motion.div variants={itemVariants} className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <Zap className="w-5 h-5 text-safety-500" />
            <span className="text-sm font-semibold text-safety-500 uppercase tracking-wide">
              Dashboard
            </span>
          </div>
          <h1 className="text-3xl sm:text-4xl font-display tracking-wide text-charcoal-900 dark:text-white mb-2">
            {greeting()}, {user?.firstName}
          </h1>
          <p className="text-charcoal-500 dark:text-charcoal-400">
            Here's what's happening with your tickets today.
          </p>
        </div>
        <Link to="/tickets/new">
          <Button size="lg" rightIcon={<ArrowRight className="w-4 h-4" />}>
            New Ticket
          </Button>
        </Link>
      </motion.div>

      {/* Stats Grid */}
      <motion.div variants={itemVariants} className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((stat, index) => (
          <motion.div
            key={stat.title}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
          >
            <Card
              variant="default"
              padding="md"
              className={cn(
                "relative overflow-hidden",
                "hover:shadow-lg transition-shadow"
              )}
            >
              <div className="flex items-start justify-between relative">
                <div>
                  <p className="text-sm font-medium text-charcoal-500 dark:text-charcoal-400 mb-1">
                    {stat.title}
                  </p>
                  <p className="text-4xl font-display text-charcoal-900 dark:text-white">
                    {stat.value}
                  </p>
                </div>
                <div className={cn(
                  "p-2.5 rounded-xl border-2",
                  stat.bgColor,
                  stat.borderColor
                )}>
                  <stat.icon className={cn("w-5 h-5", stat.color)} />
                </div>
              </div>
              {/* Subtle corner accent */}
              <div className={cn(
                "absolute -bottom-4 -right-4 w-16 h-16 rounded-full opacity-20",
                stat.bgColor
              )} />
            </Card>
          </motion.div>
        ))}
      </motion.div>

      {/* Revenue Cards */}
      <motion.div variants={itemVariants} className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Card variant="highlight" padding="lg" className="relative overflow-hidden">
          {/* Background decoration */}
          <div className="absolute top-0 right-0 w-40 h-40 bg-safety-500/10 rounded-full -translate-y-1/2 translate-x-1/2" />
          <div className="absolute bottom-0 left-0 w-20 h-20 bg-safety-500/5 rounded-full translate-y-1/2 -translate-x-1/2" />

          <div className="relative">
            <div className="flex items-center gap-3 mb-4">
              <div className="p-3 rounded-xl bg-gradient-to-br from-safety-500 to-safety-600 shadow-lg shadow-safety-500/25">
                <DollarSign className="w-6 h-6 text-white" />
              </div>
              <div>
                <span className="text-sm font-semibold text-charcoal-500 dark:text-charcoal-400 block">
                  Total Value
                </span>
                <span className="text-xs text-charcoal-400 dark:text-charcoal-500">
                  Lifetime approved
                </span>
              </div>
            </div>
            <p className="text-5xl font-display text-charcoal-900 dark:text-white mb-2">
              {formatCurrency(stats.totalValue)}
            </p>
            <div className="flex items-center gap-2 text-sm">
              <TrendingUp className="w-4 h-4 text-emerald-500" />
              <span className="text-emerald-600 dark:text-emerald-400 font-medium">+12.5%</span>
              <span className="text-charcoal-400">from last month</span>
            </div>
          </div>
        </Card>

        <Card variant="elevated" padding="lg" className="relative overflow-hidden">
          {/* Background decoration */}
          <div className="absolute top-0 right-0 w-40 h-40 bg-emerald-500/10 rounded-full -translate-y-1/2 translate-x-1/2" />

          <div className="relative">
            <div className="flex items-center gap-3 mb-4">
              <div className="p-3 rounded-xl bg-gradient-to-br from-emerald-500 to-emerald-600 shadow-lg shadow-emerald-500/25">
                <TrendingUp className="w-6 h-6 text-white" />
              </div>
              <div>
                <span className="text-sm font-semibold text-charcoal-500 dark:text-charcoal-400 block">
                  This Month
                </span>
                <span className="text-xs text-charcoal-400 dark:text-charcoal-500">
                  December 2024
                </span>
              </div>
            </div>
            <p className="text-5xl font-display text-charcoal-900 dark:text-white mb-2">
              {formatCurrency(stats.thisMonthValue)}
            </p>
            <div className="flex items-center gap-2 text-sm">
              <div className="flex -space-x-1">
                {[...Array(3)].map((_, i) => (
                  <div
                    key={i}
                    className="w-6 h-6 rounded-full bg-charcoal-300 dark:bg-charcoal-600 border-2 border-white dark:border-charcoal-900"
                  />
                ))}
              </div>
              <span className="text-charcoal-500 dark:text-charcoal-400">
                {stats.approvedTickets} tickets approved
              </span>
            </div>
          </div>
        </Card>
      </motion.div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Tickets */}
        <motion.div variants={itemVariants} className="lg:col-span-2">
          <Card variant="default" padding="none">
            <CardHeader className="p-5 border-b-2 border-charcoal-100 dark:border-charcoal-800">
              <CardTitle>Recent Tickets</CardTitle>
              <Link to="/tickets">
                <Button variant="ghost" size="sm" rightIcon={<ArrowRight className="w-4 h-4" />}>
                  View All
                </Button>
              </Link>
            </CardHeader>
            <div className="divide-y-2 divide-charcoal-100 dark:divide-charcoal-800">
              {recentTickets.map((ticket, index) => (
                <Link
                  key={ticket.id}
                  to={`/tickets/${ticket.id}`}
                  className={cn(
                    "block p-5",
                    "hover:bg-charcoal-50 dark:hover:bg-charcoal-800/50",
                    "transition-colors"
                  )}
                >
                  <motion.div
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: index * 0.05 }}
                    className="flex items-start justify-between gap-4"
                  >
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1.5">
                        <span className="font-bold text-safety-500">{ticket.ticketNumber}</span>
                        <StatusBadge status={ticket.status} size="sm" />
                      </div>
                      <p className="font-semibold text-charcoal-900 dark:text-white truncate mb-1">
                        {ticket.title}
                      </p>
                      <p className="text-sm text-charcoal-500 dark:text-charcoal-400 truncate">
                        {ticket.project?.name}
                      </p>
                    </div>
                    <div className="text-right shrink-0">
                      <p className="font-bold text-charcoal-900 dark:text-white">
                        {formatCurrency(ticket.total)}
                      </p>
                      <p className="text-sm text-charcoal-500 dark:text-charcoal-400 flex items-center gap-1.5 justify-end mt-1">
                        <CalendarDays className="w-3.5 h-3.5" />
                        {ticket.workDate}
                      </p>
                    </div>
                  </motion.div>
                </Link>
              ))}
            </div>
          </Card>
        </motion.div>

        {/* Activity Feed */}
        <motion.div variants={itemVariants}>
          <Card variant="default" padding="none">
            <CardHeader className="p-5 border-b-2 border-charcoal-100 dark:border-charcoal-800">
              <CardTitle>Activity</CardTitle>
            </CardHeader>
            <div className="p-5 space-y-4">
              {stats.recentActivity.map((activity, index) => (
                <motion.div
                  key={activity.id}
                  className="flex gap-3"
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.05 }}
                >
                  <div className="shrink-0 mt-0.5 p-1.5 rounded-lg bg-charcoal-100 dark:bg-charcoal-800">
                    {getActivityIcon(activity.type)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm">
                      <span className="font-bold text-charcoal-900 dark:text-white">
                        {activity.ticketNumber}
                      </span>
                      {' '}
                      <span className="text-charcoal-500 dark:text-charcoal-400">
                        {activity.message}
                      </span>
                    </p>
                    <p className="text-xs text-charcoal-400 dark:text-charcoal-500 mt-1">
                      {formatRelativeTime(activity.timestamp)}
                    </p>
                  </div>
                </motion.div>
              ))}
            </div>
          </Card>
        </motion.div>
      </div>
    </motion.div>
  );
}
