import { useState } from 'react';
import { motion } from 'framer-motion';
import { useNavigate, useParams } from 'react-router-dom';
import {
  ArrowLeft,
  Edit2,
  Send,
  FileText,
  Calendar,
  MapPin,
  User,
  Clock,
  CheckCircle2,
  XCircle,
  Lock,
  AlertCircle,
  Camera,
  Signature,
  DollarSign,
} from 'lucide-react';
import { Card, Button, StatusBadge, Badge, Modal } from '@/components/ui';
import { SignatureCapture } from '@/components/signature/SignatureCapture';
import { useAppStore } from '@/stores/app';
import { formatCurrency, formatDate, formatDateTime } from '@/lib/utils';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

export function TicketDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { tickets, updateTicket, user } = useAppStore();
  const [showSignatureModal, setShowSignatureModal] = useState(false);
  const [selectedPhoto, setSelectedPhoto] = useState<string | null>(null);

  const ticket = tickets.find((t) => t.id === id);

  if (!ticket) {
    return (
      <div className="text-center py-12">
        <h2 className="text-2xl mb-2">Ticket Not Found</h2>
        <Button onClick={() => navigate('/tickets')}>Back to Tickets</Button>
      </div>
    );
  }

  const canEdit = ticket.status === 'draft' || ticket.status === 'rejected';
  const canSubmit = ticket.status === 'draft' && ticket.lineItems.length > 0;
  const canSign = ticket.status === 'pending_review' && !ticket.gcSignature;

  const handleSubmit = () => {
    updateTicket(ticket.id, { status: 'pending_review', submittedAt: new Date().toISOString() });
  };

  const handleSignatureComplete = (signatures: {
    foremanSignature: string;
    gcSignature?: string;
    gcSignerName?: string;
    signingMethod: 'link_own_device' | 'same_device_verified';
  }) => {
    const now = new Date().toISOString();

    // Create foreman signature
    const foremanSig = {
      id: `sig-foreman-${Date.now()}`,
      ticketId: ticket.id,
      signerType: 'foreman' as const,
      signerName: `${user?.firstName} ${user?.lastName}`,
      signerEmail: user?.email || '',
      signatureImageUrl: signatures.foremanSignature,
      signingMethod: 'same_device_verified' as const,
      signedAt: now,
    };

    // If GC signed on-device, create their signature and approve
    if (signatures.gcSignature && signatures.gcSignerName) {
      const gcSig = {
        id: `sig-gc-${Date.now()}`,
        ticketId: ticket.id,
        signerType: 'gc_super' as const,
        signerName: signatures.gcSignerName,
        signerEmail: ticket.project?.gcContactEmail || '',
        signatureImageUrl: signatures.gcSignature,
        signingMethod: signatures.signingMethod,
        signedAt: now,
      };

      updateTicket(ticket.id, {
        status: 'approved',
        approvedAt: now,
        foremanSignature: foremanSig,
        gcSignature: gcSig,
      });
    } else {
      // GC will sign later via link - submit for review
      updateTicket(ticket.id, {
        status: 'pending_review',
        submittedAt: now,
        foremanSignature: foremanSig,
      });
    }

    setShowSignatureModal(false);
  };

  const getStatusTimeline = () => {
    const events = [];
    events.push({ label: 'Created', date: ticket.createdAt, icon: FileText });
    if (ticket.submittedAt) {
      events.push({ label: 'Submitted', date: ticket.submittedAt, icon: Send });
    }
    if (ticket.approvedAt) {
      events.push({ label: 'Approved', date: ticket.approvedAt, icon: CheckCircle2 });
    }
    if (ticket.rejectedAt) {
      events.push({ label: 'Rejected', date: ticket.rejectedAt, icon: XCircle });
    }
    if (ticket.lockedAt) {
      events.push({ label: 'Locked', date: ticket.lockedAt, icon: Lock });
    }
    return events;
  };

  return (
    <motion.div variants={containerVariants} initial="hidden" animate="visible" className="space-y-6 max-w-4xl mx-auto">
      {/* Header */}
      <motion.div variants={itemVariants} className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="default" onClick={() => navigate(-1)} leftIcon={<ArrowLeft className="w-5 h-5" />}>
            Back
          </Button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl sm:text-3xl">{ticket.ticketNumber}</h1>
              <StatusBadge status={ticket.status} size="lg" />
            </div>
            <p className="text-secondary">{ticket.project?.name}</p>
          </div>
        </div>
        <div className="flex gap-2">
          {canEdit && (
            <Button variant="outline" onClick={() => navigate(`/tickets/${ticket.id}/edit`)} leftIcon={<Edit2 className="w-4 h-4" />}>
              Edit
            </Button>
          )}
          {canSubmit && (
            <Button onClick={handleSubmit} leftIcon={<Send className="w-4 h-4" />}>
              Submit
            </Button>
          )}
          {canSign && (
            <Button onClick={() => setShowSignatureModal(true)} leftIcon={<Signature className="w-4 h-4" />}>
              Get Signature
            </Button>
          )}
        </div>
      </motion.div>

      {/* Rejection Notice */}
      {ticket.status === 'rejected' && ticket.rejectionReason && (
        <motion.div variants={itemVariants}>
          <Card variant="default" padding="md" className="border-2 border-red-300 dark:border-red-800 bg-red-50 dark:bg-red-900/20">
            <div className="flex gap-3">
              <AlertCircle className="w-5 h-5 text-red-500 shrink-0 mt-0.5" />
              <div>
                <p className="font-medium text-red-800 dark:text-red-200">Ticket Rejected</p>
                <p className="text-red-700 dark:text-red-300 mt-1">{ticket.rejectionReason}</p>
              </div>
            </div>
          </Card>
        </motion.div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Title & Description */}
          <motion.div variants={itemVariants}>
            <Card variant="default" padding="lg">
              <h2 className="text-xl mb-4">{ticket.title}</h2>
              {ticket.descriptionCleaned || ticket.descriptionRaw ? (
                <p className="text-secondary whitespace-pre-wrap">
                  {ticket.descriptionCleaned || ticket.descriptionRaw}
                </p>
              ) : (
                <p className="text-tertiary italic">No description provided.</p>
              )}
            </Card>
          </motion.div>

          {/* Line Items */}
          <motion.div variants={itemVariants}>
            <Card variant="default" padding="none">
              <div className="p-4 sm:p-5 border-b border-default">
                <h2 className="text-xl">Line Items</h2>
              </div>
              <div className="divide-y divide-charcoal-100 dark:divide-charcoal-800">
                {ticket.lineItems.map((item) => (
                  <div key={item.id} className="p-4 sm:p-5 flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <Badge variant={item.itemType === 'labor' ? 'info' : item.itemType === 'material' ? 'success' : 'warning'} size="sm">
                          {item.itemType}
                        </Badge>
                      </div>
                      <p className="font-medium">{item.description}</p>
                      <p className="text-sm text-secondary">
                        {item.quantity} {item.unit} × {formatCurrency(item.unitPrice)}
                      </p>
                    </div>
                    <p className="font-display text-xl">{formatCurrency(item.total)}</p>
                  </div>
                ))}
              </div>
              <div className="p-4 sm:p-5 bg-charcoal-50 dark:bg-charcoal-800/50 flex items-center justify-between">
                <span className="font-medium text-lg">Total</span>
                <span className="font-display text-3xl text-safety-500">{formatCurrency(ticket.total)}</span>
              </div>
            </Card>
          </motion.div>

          {/* Photos */}
          {ticket.photos.length > 0 && (
            <motion.div variants={itemVariants}>
              <Card variant="default" padding="lg">
                <h2 className="text-xl mb-4 flex items-center gap-2">
                  <Camera className="w-5 h-5" />
                  Photos ({ticket.photos.length})
                </h2>
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                  {ticket.photos.map((photo) => (
                    <button
                      key={photo.id}
                      onClick={() => setSelectedPhoto(photo.fileUrl)}
                      className="aspect-square rounded-lg overflow-hidden bg-charcoal-100 dark:bg-charcoal-800 hover:ring-2 ring-safety-500 transition-all"
                    >
                      <img src={photo.thumbnailUrl} alt={photo.caption || ''} className="w-full h-full object-cover" />
                    </button>
                  ))}
                </div>
              </Card>
            </motion.div>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Details */}
          <motion.div variants={itemVariants}>
            <Card variant="default" padding="lg">
              <h3 className="text-lg mb-4">Details</h3>
              <div className="space-y-4">
                <div className="flex items-start gap-3">
                  <Calendar className="w-4 h-4 text-charcoal-400 mt-0.5" />
                  <div>
                    <p className="text-sm text-secondary">Work Date</p>
                    <p className="font-medium">{formatDate(ticket.workDate)}</p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <MapPin className="w-4 h-4 text-charcoal-400 mt-0.5" />
                  <div>
                    <p className="text-sm text-secondary">Project</p>
                    <p className="font-medium">{ticket.project?.name}</p>
                    <p className="text-sm text-tertiary">{ticket.project?.address}</p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <User className="w-4 h-4 text-charcoal-400 mt-0.5" />
                  <div>
                    <p className="text-sm text-secondary">Created By</p>
                    <p className="font-medium">{ticket.createdBy?.firstName} {ticket.createdBy?.lastName}</p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <DollarSign className="w-4 h-4 text-charcoal-400 mt-0.5" />
                  <div>
                    <p className="text-sm text-secondary">Ticket Type</p>
                    <p className="font-medium capitalize">{ticket.ticketType.replace('_', ' ')}</p>
                  </div>
                </div>
              </div>
            </Card>
          </motion.div>

          {/* Signatures */}
          <motion.div variants={itemVariants}>
            <Card variant="default" padding="lg">
              <h3 className="text-lg mb-4 flex items-center gap-2">
                <Signature className="w-5 h-5" />
                Signatures
              </h3>
              <div className="space-y-4">
                {ticket.foremanSignature ? (
                  <div className="p-3 bg-emerald-50 dark:bg-emerald-900/20 rounded-lg">
                    <p className="text-sm text-emerald-700 dark:text-emerald-400 font-medium">Foreman</p>
                    <p className="font-medium">{ticket.foremanSignature.signerName}</p>
                    <p className="text-xs text-secondary">{formatDateTime(ticket.foremanSignature.signedAt)}</p>
                  </div>
                ) : (
                  <div className="p-3 bg-charcoal-50 dark:bg-charcoal-800 rounded-lg">
                    <p className="text-sm text-secondary">Foreman signature pending</p>
                  </div>
                )}

                {ticket.gcSignature ? (
                  <div className="p-3 bg-emerald-50 dark:bg-emerald-900/20 rounded-lg">
                    <p className="text-sm text-emerald-700 dark:text-emerald-400 font-medium">GC Superintendent</p>
                    <p className="font-medium">{ticket.gcSignature.signerName}</p>
                    <p className="text-xs text-secondary">{formatDateTime(ticket.gcSignature.signedAt)}</p>
                  </div>
                ) : (
                  <div className="p-3 bg-charcoal-50 dark:bg-charcoal-800 rounded-lg">
                    <p className="text-sm text-secondary">GC signature pending</p>
                    {canSign && (
                      <Button variant="outline" size="sm" className="mt-2" onClick={() => setShowSignatureModal(true)}>
                        Collect Signature
                      </Button>
                    )}
                  </div>
                )}
              </div>
            </Card>
          </motion.div>

          {/* Timeline */}
          <motion.div variants={itemVariants}>
            <Card variant="default" padding="lg">
              <h3 className="text-lg mb-4 flex items-center gap-2">
                <Clock className="w-5 h-5" />
                Timeline
              </h3>
              <div className="space-y-3">
                {getStatusTimeline().map((event, idx) => (
                  <div key={idx} className="flex items-start gap-3">
                    <div className="p-1.5 rounded-full bg-charcoal-100 dark:bg-charcoal-800">
                      <event.icon className="w-3 h-3 text-charcoal-500" />
                    </div>
                    <div>
                      <p className="text-sm font-medium">{event.label}</p>
                      <p className="text-xs text-secondary">{formatDateTime(event.date)}</p>
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          </motion.div>
        </div>
      </div>

      {/* Signature Modal */}
      <Modal
        isOpen={showSignatureModal}
        onClose={() => setShowSignatureModal(false)}
        title="Collect GC Signature"
        size="lg"
      >
        {ticket.project && (
          <SignatureCapture
            project={ticket.project}
            foremanName={`${user?.firstName} ${user?.lastName}`}
            onComplete={handleSignatureComplete}
            onCancel={() => setShowSignatureModal(false)}
          />
        )}
      </Modal>

      {/* Photo Modal */}
      <Modal
        isOpen={!!selectedPhoto}
        onClose={() => setSelectedPhoto(null)}
        title="Photo"
        size="xl"
      >
        {selectedPhoto && (
          <img src={selectedPhoto} alt="" className="w-full rounded-lg" />
        )}
      </Modal>
    </motion.div>
  );
}
