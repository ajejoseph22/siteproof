import { motion } from 'framer-motion';
import {
  Building2,
  User,
  Bell,
  Shield,
  Palette,
  Moon,
  Sun,
  ChevronRight,
  Mail,
  Phone,
  MapPin,
  Camera,
} from 'lucide-react';
import { Card, Button, Avatar } from '@/components/ui';
import { useAppStore } from '@/stores/app';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

export function SettingsPage() {
  const { user, company, isDarkMode, toggleDarkMode } = useAppStore();

  const settingsSections = [
    {
      title: 'Notifications',
      icon: Bell,
      items: [
        { label: 'Email notifications', value: 'Enabled', toggle: true },
        { label: 'SMS alerts', value: 'Critical only', toggle: false },
        { label: 'Push notifications', value: 'All activity', toggle: false },
      ],
    },
    {
      title: 'Security',
      icon: Shield,
      items: [
        { label: 'Two-factor authentication', value: 'Enabled', toggle: true },
        { label: 'Session timeout', value: '30 minutes', toggle: false },
        { label: 'Password', value: 'Last changed 30 days ago', toggle: false },
      ],
    },
  ];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className="space-y-6 max-w-4xl mx-auto"
    >
      {/* Header */}
      <motion.div variants={itemVariants}>
        <h1 className="text-3xl sm:text-4xl mb-2">Settings</h1>
        <p className="text-secondary">Manage your account and preferences.</p>
      </motion.div>

      {/* Profile Section */}
      <motion.div variants={itemVariants}>
        <Card variant="default" padding="lg">
          <div className="flex items-start gap-4 sm:gap-6">
            <div className="relative">
              <Avatar
                name={`${user?.firstName} ${user?.lastName}`}
                src={user?.avatarUrl}
                size="xl"
              />
              <button className="absolute -bottom-1 -right-1 p-1.5 rounded-full bg-safety-500 text-white shadow-lg hover:bg-safety-600 transition-colors">
                <Camera className="w-4 h-4" />
              </button>
            </div>
            <div className="flex-1">
              <h2 className="text-2xl mb-1">
                {user?.firstName} {user?.lastName}
              </h2>
              <p className="text-secondary capitalize">{user?.role?.replace('_', ' ')}</p>
              <div className="flex flex-wrap gap-4 mt-3 text-sm text-secondary">
                <span className="flex items-center gap-1.5">
                  <Mail className="w-4 h-4" />
                  {user?.email}
                </span>
              </div>
            </div>
            <Button variant="outline" size="sm">
              Edit Profile
            </Button>
          </div>
        </Card>
      </motion.div>

      {/* Company Section */}
      <motion.div variants={itemVariants}>
        <Card variant="default" padding="lg">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <Building2 className="w-5 h-5 text-charcoal-400" />
              <h2 className="text-xl">Company</h2>
            </div>
            <Button variant="outline" size="sm">
              Edit
            </Button>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div className="flex items-start gap-3">
              <div className="w-12 h-12 rounded-lg bg-safety-500 flex items-center justify-center text-white font-display text-lg">
                {company?.name?.slice(0, 2).toUpperCase()}
              </div>
              <div>
                <p className="font-medium">{company?.name}</p>
                <p className="text-sm text-secondary">Company Name</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <MapPin className="w-5 h-5 text-charcoal-400 mt-0.5" />
              <div>
                <p className="font-medium">{company?.address}</p>
                <p className="text-sm text-secondary">Address</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <Phone className="w-5 h-5 text-charcoal-400 mt-0.5" />
              <div>
                <p className="font-medium">{company?.phone}</p>
                <p className="text-sm text-secondary">Phone</p>
              </div>
            </div>
          </div>
        </Card>
      </motion.div>

      {/* Appearance */}
      <motion.div variants={itemVariants}>
        <Card variant="default" padding="lg">
          <div className="flex items-center gap-3 mb-6">
            <Palette className="w-5 h-5 text-charcoal-400" />
            <h2 className="text-xl">Appearance</h2>
          </div>

          <div className="flex items-center justify-between p-4 bg-charcoal-50 dark:bg-charcoal-800/50 rounded-lg">
            <div className="flex items-center gap-3">
              {isDarkMode ? (
                <Moon className="w-5 h-5 text-charcoal-400" />
              ) : (
                <Sun className="w-5 h-5 text-charcoal-400" />
              )}
              <div>
                <p className="font-medium">Theme</p>
                <p className="text-sm text-secondary">
                  {isDarkMode ? 'Dark mode' : 'Light mode'}
                </p>
              </div>
            </div>
            <Button
              variant="outline"
              onClick={toggleDarkMode}
              leftIcon={isDarkMode ? <Sun className="w-4 h-4" /> : <Moon className="w-4 h-4" />}
            >
              {isDarkMode ? 'Light' : 'Dark'}
            </Button>
          </div>
        </Card>
      </motion.div>

      {/* Other Settings */}
      {settingsSections.map((section) => (
        <motion.div key={section.title} variants={itemVariants}>
          <Card variant="default" padding="lg">
            <div className="flex items-center gap-3 mb-6">
              <section.icon className="w-5 h-5 text-charcoal-400" />
              <h2 className="text-xl">{section.title}</h2>
            </div>

            <div className="divide-y divide-charcoal-100 dark:divide-charcoal-800 -mx-5 sm:-mx-6">
              {section.items.map((item, idx) => (
                <div
                  key={idx}
                  className="px-5 sm:px-6 py-4 flex items-center justify-between hover:bg-charcoal-50 dark:hover:bg-charcoal-800/50 transition-colors cursor-pointer"
                >
                  <div>
                    <p className="font-medium">{item.label}</p>
                    <p className="text-sm text-secondary">{item.value}</p>
                  </div>
                  {item.toggle ? (
                    <button
                      className="relative w-11 h-6 bg-safety-500 rounded-full transition-colors"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <span className="absolute right-1 top-1 w-4 h-4 bg-white rounded-full shadow transition-transform" />
                    </button>
                  ) : (
                    <ChevronRight className="w-5 h-5 text-charcoal-400" />
                  )}
                </div>
              ))}
            </div>
          </Card>
        </motion.div>
      ))}

      {/* Team Members */}
      <motion.div variants={itemVariants}>
        <Card variant="default" padding="lg">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <User className="w-5 h-5 text-charcoal-400" />
              <h2 className="text-xl">Team Members</h2>
            </div>
            <Button variant="outline" size="sm">
              Invite
            </Button>
          </div>

          <div className="space-y-3">
            {[
              { name: 'Mike Rodriguez', email: 'mike.rodriguez@apexelectric.com', role: 'Foreman' },
              { name: 'Sarah Johnson', email: 'sarah.johnson@apexelectric.com', role: 'Project Manager' },
              { name: 'Tom Williams', email: 'tom.williams@apexelectric.com', role: 'Admin' },
            ].map((member, idx) => (
              <div
                key={idx}
                className="flex items-center gap-4 p-3 rounded-lg hover:bg-charcoal-50 dark:hover:bg-charcoal-800/50 transition-colors"
              >
                <Avatar name={member.name} size="md" />
                <div className="flex-1 min-w-0">
                  <p className="font-medium truncate">{member.name}</p>
                  <p className="text-sm text-secondary truncate">{member.email}</p>
                </div>
                <span className="text-sm text-secondary">{member.role}</span>
              </div>
            ))}
          </div>
        </Card>
      </motion.div>

      {/* Danger Zone */}
      <motion.div variants={itemVariants}>
        <Card variant="outline" padding="lg" className="border-red-200 dark:border-red-800">
          <h2 className="text-xl text-red-600 dark:text-red-400 mb-4">Danger Zone</h2>
          <p className="text-secondary mb-4">
            Once you delete your account, there is no going back. Please be certain.
          </p>
          <Button variant="destructive" size="sm">
            Delete Account
          </Button>
        </Card>
      </motion.div>
    </motion.div>
  );
}
