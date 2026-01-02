import { useState, useMemo } from 'react';
import { motion } from 'framer-motion';
import { Link, useSearchParams } from 'react-router-dom';
import {
  Search,
  Plus,
  Filter,
  CalendarDays,
  Camera,
  Signature,
  ChevronRight,
  ClipboardList,
  SlidersHorizontal,
} from 'lucide-react';
import { Card, Input, Button, StatusBadge, EmptyState } from '@/components/ui';
import { useAppStore } from '@/stores/app';
import { formatCurrency, formatDate } from '@/lib/utils';
import type { TicketStatus } from '@/types';

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

const statusFilters: { value: TicketStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'draft', label: 'Draft' },
  { value: 'pending_review', label: 'Pending' },
  { value: 'approved', label: 'Approved' },
  { value: 'rejected', label: 'Rejected' },
  { value: 'locked', label: 'Locked' },
];

export function TicketsPage() {
  const { tickets, projects } = useAppStore();
  const [searchParams, setSearchParams] = useSearchParams();
  const [search, setSearch] = useState('');
  const [showFilters, setShowFilters] = useState(false);

  const statusFilter = (searchParams.get('status') as TicketStatus | 'all') || 'all';
  const projectFilter = searchParams.get('project') || 'all';

  const filteredTickets = useMemo(() => {
    return tickets.filter((ticket) => {
      const matchesSearch =
        ticket.ticketNumber.toLowerCase().includes(search.toLowerCase()) ||
        ticket.title.toLowerCase().includes(search.toLowerCase()) ||
        ticket.project?.name.toLowerCase().includes(search.toLowerCase());

      const matchesStatus = statusFilter === 'all' || ticket.status === statusFilter;
      const matchesProject = projectFilter === 'all' || ticket.projectId === projectFilter;

      return matchesSearch && matchesStatus && matchesProject;
    });
  }, [tickets, search, statusFilter, projectFilter]);

  const setFilter = (key: string, value: string) => {
    const params = new URLSearchParams(searchParams);
    if (value === 'all') {
      params.delete(key);
    } else {
      params.set(key, value);
    }
    setSearchParams(params);
  };

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className="space-y-6"
    >
      {/* Header */}
      <motion.div variants={itemVariants} className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl sm:text-4xl mb-2">Tickets</h1>
          <p className="text-secondary">
            {filteredTickets.length} ticket{filteredTickets.length !== 1 ? 's' : ''} found
          </p>
        </div>
        <Link to="/tickets/new">
          <Button size="lg" leftIcon={<Plus className="w-5 h-5" />}>
            New Ticket
          </Button>
        </Link>
      </motion.div>

      {/* Search & Filters */}
      <motion.div variants={itemVariants} className="space-y-4">
        <div className="flex gap-3">
          <div className="flex-1">
            <Input
              placeholder="Search tickets..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              leftIcon={<Search className="w-5 h-5" />}
            />
          </div>
          <Button
            variant="outline"
            onClick={() => setShowFilters(!showFilters)}
            leftIcon={<SlidersHorizontal className="w-5 h-5" />}
            className={showFilters ? 'border-safety-500 text-safety-500' : ''}
          >
            Filters
          </Button>
        </div>

        {/* Status Filters - Always visible */}
        <div className="flex gap-2 overflow-x-auto pb-2">
          {statusFilters.map((filter) => (
            <Button
              key={filter.value}
              variant={statusFilter === filter.value ? 'default' : 'outline'}
              size="sm"
              onClick={() => setFilter('status', filter.value)}
            >
              {filter.label}
            </Button>
          ))}
        </div>

        {/* Extended Filters */}
        {showFilters && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="p-4 bg-charcoal-50 dark:bg-charcoal-800/50 rounded-lg"
          >
            <div className="flex items-center gap-2 mb-3">
              <Filter className="w-4 h-4 text-secondary" />
              <span className="text-sm font-medium">Filter by Project</span>
            </div>
            <div className="flex flex-wrap gap-2">
              <Button
                variant={projectFilter === 'all' ? 'secondary' : 'ghost'}
                size="sm"
                onClick={() => setFilter('project', 'all')}
              >
                All Projects
              </Button>
              {projects.map((project) => (
                <Button
                  key={project.id}
                  variant={projectFilter === project.id ? 'secondary' : 'ghost'}
                  size="sm"
                  onClick={() => setFilter('project', project.id)}
                >
                  {project.name}
                </Button>
              ))}
            </div>
          </motion.div>
        )}
      </motion.div>

      {/* Tickets List */}
      <motion.div variants={itemVariants} className="space-y-3">
        {filteredTickets.map((ticket) => (
          <motion.div key={ticket.id} variants={itemVariants}>
            <Card variant="default" hover padding="none">
              <Link to={`/tickets/${ticket.id}`} className="block p-4 sm:p-5">
                <div className="flex items-start gap-4">
                  {/* Main Content */}
                  <div className="flex-1 min-w-0">
                    {/* Top Row */}
                    <div className="flex items-center gap-2 flex-wrap mb-2">
                      <span className="font-medium text-safety-500">{ticket.ticketNumber}</span>
                      <StatusBadge status={ticket.status} />
                      <span className="text-sm text-tertiary">•</span>
                      <span className="text-sm text-secondary">{ticket.project?.name}</span>
                    </div>

                    {/* Title */}
                    <h3 className="font-medium text-lg mb-2 truncate">{ticket.title}</h3>

                    {/* Meta Row */}
                    <div className="flex items-center gap-4 text-sm text-secondary flex-wrap">
                      <div className="flex items-center gap-1.5">
                        <CalendarDays className="w-4 h-4" />
                        <span>{formatDate(ticket.workDate)}</span>
                      </div>
                      <div className="flex items-center gap-1.5">
                        <Camera className="w-4 h-4" />
                        <span>{ticket.photos.length} photos</span>
                      </div>
                      {(ticket.foremanSignature || ticket.gcSignature) && (
                        <div className="flex items-center gap-1.5">
                          <Signature className="w-4 h-4" />
                          <span>
                            {[ticket.foremanSignature && 'Foreman', ticket.gcSignature && 'GC']
                              .filter(Boolean)
                              .join(', ')}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Right Side */}
                  <div className="flex items-center gap-4 shrink-0">
                    <div className="text-right hidden sm:block">
                      <p className="font-display text-xl">{formatCurrency(ticket.total)}</p>
                      <p className="text-xs text-secondary uppercase tracking-wide">
                        {ticket.lineItems.length} items
                      </p>
                    </div>
                    <ChevronRight className="w-5 h-5 text-charcoal-400" />
                  </div>
                </div>

                {/* Mobile Price */}
                <div className="sm:hidden mt-3 pt-3 border-t border-default flex items-center justify-between">
                  <span className="text-sm text-secondary">{ticket.lineItems.length} line items</span>
                  <span className="font-display text-xl">{formatCurrency(ticket.total)}</span>
                </div>
              </Link>
            </Card>
          </motion.div>
        ))}
      </motion.div>

      {/* Empty State */}
      {filteredTickets.length === 0 && (
        <EmptyState
          icon={<ClipboardList className="w-8 h-8" />}
          title="No Tickets Found"
          description={
            search || statusFilter !== 'all' || projectFilter !== 'all'
              ? 'Try adjusting your search or filters.'
              : 'Create your first T&M ticket to get started.'
          }
          action={
            <Link to="/tickets/new">
              <Button leftIcon={<Plus className="w-5 h-5" />}>Create Ticket</Button>
            </Link>
          }
        />
      )}
    </motion.div>
  );
}
