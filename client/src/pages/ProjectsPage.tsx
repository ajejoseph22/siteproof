import { useState } from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import {
  Search,
  MapPin,
  Building2,
  Phone,
  Mail,
  ClipboardList,
  DollarSign,
  ChevronRight,
} from 'lucide-react';
import { Card, Input, Badge, Button } from '@/components/ui';
import { useAppStore } from '@/stores/app';
import { formatCurrency } from '@/lib/utils';
import type { Project } from '@/types';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.05 },
  },
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

export function ProjectsPage() {
  const { projects } = useAppStore();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');

  const filteredProjects = projects.filter((project) => {
    const matchesSearch =
      project.name.toLowerCase().includes(search.toLowerCase()) ||
      project.address.toLowerCase().includes(search.toLowerCase()) ||
      project.gcCompanyName.toLowerCase().includes(search.toLowerCase());

    const matchesStatus = statusFilter === 'all' || project.status === statusFilter;

    return matchesSearch && matchesStatus;
  });

  const getStatusBadge = (status: Project['status']) => {
    switch (status) {
      case 'active':
        return <Badge variant="success">Active</Badge>;
      case 'completed':
        return <Badge variant="info">Completed</Badge>;
      case 'on_hold':
        return <Badge variant="warning">On Hold</Badge>;
      default:
        return <Badge>{status}</Badge>;
    }
  };

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className="space-y-6"
    >
      {/* Header */}
      <motion.div variants={itemVariants}>
        <h1 className="text-3xl sm:text-4xl mb-2">Projects</h1>
        <p className="text-secondary">Manage your active construction projects.</p>
      </motion.div>

      {/* Filters */}
      <motion.div variants={itemVariants} className="flex flex-col sm:flex-row gap-4">
        <div className="flex-1">
          <Input
            placeholder="Search projects..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            leftIcon={<Search className="w-5 h-5" />}
          />
        </div>
        <div className="flex gap-2">
          {['all', 'active', 'on_hold', 'completed'].map((status) => (
            <Button
              key={status}
              variant={statusFilter === status ? 'default' : 'outline'}
              size="default"
              onClick={() => setStatusFilter(status)}
            >
              {status === 'all' ? 'All' : status === 'on_hold' ? 'On Hold' : status.charAt(0).toUpperCase() + status.slice(1)}
            </Button>
          ))}
        </div>
      </motion.div>

      {/* Projects Grid */}
      <motion.div variants={itemVariants} className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {filteredProjects.map((project) => (
          <motion.div key={project.id} variants={itemVariants}>
            <Card variant="default" hover padding="none" className="overflow-hidden">
              <Link to={`/tickets?project=${project.id}`} className="block">
                {/* Header with status */}
                <div className="p-4 sm:p-5 border-b border-default">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        {getStatusBadge(project.status)}
                      </div>
                      <h3 className="font-display text-xl tracking-wide truncate">
                        {project.name}
                      </h3>
                    </div>
                    <ChevronRight className="w-5 h-5 text-charcoal-400 shrink-0" />
                  </div>
                </div>

                {/* Details */}
                <div className="p-4 sm:p-5 space-y-3">
                  <div className="flex items-start gap-3">
                    <MapPin className="w-4 h-4 text-charcoal-400 shrink-0 mt-0.5" />
                    <span className="text-sm text-secondary">{project.address}</span>
                  </div>
                  <div className="flex items-start gap-3">
                    <Building2 className="w-4 h-4 text-charcoal-400 shrink-0 mt-0.5" />
                    <div>
                      <span className="text-sm font-medium">{project.gcCompanyName}</span>
                      <span className="text-sm text-secondary block">{project.gcContactName}</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    {project.gcContactPhone && (
                      <div className="flex items-center gap-2">
                        <Phone className="w-4 h-4 text-charcoal-400" />
                        <span className="text-sm text-secondary">{project.gcContactPhone}</span>
                      </div>
                    )}
                    <div className="flex items-center gap-2">
                      <Mail className="w-4 h-4 text-charcoal-400" />
                      <span className="text-sm text-secondary truncate">{project.gcContactEmail}</span>
                    </div>
                  </div>
                </div>

                {/* Stats Footer */}
                <div className="px-4 sm:px-5 py-3 bg-charcoal-50 dark:bg-charcoal-800/50 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <ClipboardList className="w-4 h-4 text-charcoal-400" />
                    <span className="text-sm font-medium">{project.ticketCount} tickets</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <DollarSign className="w-4 h-4 text-safety-500" />
                    <span className="text-sm font-medium text-safety-600 dark:text-safety-400">
                      {formatCurrency(project.totalValue)}
                    </span>
                  </div>
                </div>
              </Link>
            </Card>
          </motion.div>
        ))}
      </motion.div>

      {/* Empty State */}
      {filteredProjects.length === 0 && (
        <Card variant="ghost" padding="lg" className="text-center">
          <Building2 className="w-12 h-12 mx-auto text-charcoal-300 dark:text-charcoal-600 mb-4" />
          <h3 className="text-xl mb-2">No Projects Found</h3>
          <p className="text-secondary">
            {search ? 'Try adjusting your search terms.' : 'No projects match the selected filters.'}
          </p>
        </Card>
      )}
    </motion.div>
  );
}
