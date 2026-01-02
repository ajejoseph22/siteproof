import { useState, useRef } from 'react';
import { motion } from 'framer-motion';
import { useNavigate, useParams } from 'react-router-dom';
import { useForm, useFieldArray } from 'react-hook-form';
import {
  ArrowLeft,
  Mic,
  MicOff,
  Camera,
  Plus,
  Trash2,
  Sparkles,
  X,
  ImageIcon,
  Save,
  Send,
} from 'lucide-react';
import { Card, Button, Input, Select, Textarea, Badge } from '@/components/ui';
import { useAppStore } from '@/stores/app';
import { formatCurrency, generateId } from '@/lib/utils';
import type { LineItemType, Ticket } from '@/types';

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

interface TicketFormData {
  projectId: string;
  title: string;
  workDate: string;
  descriptionRaw: string;
  lineItems: {
    itemType: LineItemType;
    description: string;
    quantity: number;
    unit: string;
    unitPrice: number;
  }[];
}

const unitOptions = [
  { value: 'hours', label: 'Hours' },
  { value: 'ea', label: 'Each' },
  { value: 'ft', label: 'Feet' },
  { value: 'lot', label: 'Lot' },
  { value: 'sf', label: 'Sq. Ft.' },
  { value: 'lf', label: 'Lin. Ft.' },
];

const itemTypeOptions = [
  { value: 'labor', label: 'Labor' },
  { value: 'material', label: 'Material' },
  { value: 'equipment', label: 'Equipment' },
];

export function TicketFormPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { projects, tickets, addTicket, updateTicket } = useAppStore();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const isEdit = Boolean(id);
  const existingTicket = tickets.find((t) => t.id === id);

  const [isRecording, setIsRecording] = useState(false);
  const [isCleaningUp, setIsCleaningUp] = useState(false);
  const [photos, setPhotos] = useState<{ id: string; url: string; caption: string }[]>(
    existingTicket?.photos.map((p) => ({ id: p.id, url: p.thumbnailUrl, caption: p.caption || '' })) || []
  );

  const {
    register,
    control,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<TicketFormData>({
    defaultValues: {
      projectId: existingTicket?.projectId || projects[0]?.id || '',
      title: existingTicket?.title || '',
      workDate: existingTicket?.workDate || new Date().toISOString().split('T')[0],
      descriptionRaw: existingTicket?.descriptionRaw || '',
      lineItems: existingTicket?.lineItems.map((li) => ({
        itemType: li.itemType,
        description: li.description,
        quantity: li.quantity,
        unit: li.unit,
        unitPrice: li.unitPrice,
      })) || [
        { itemType: 'labor' as LineItemType, description: '', quantity: 1, unit: 'hours', unitPrice: 85 },
      ],
    },
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: 'lineItems',
  });

  const watchLineItems = watch('lineItems');
  const total = watchLineItems.reduce((sum, item) => sum + (item.quantity || 0) * (item.unitPrice || 0), 0);

  const handleVoiceRecord = () => {
    setIsRecording(!isRecording);
    // Simulate voice recording
    if (!isRecording) {
      setTimeout(() => {
        setIsRecording(false);
        setValue(
          'descriptionRaw',
          'We had to run additional conduit because the architect changed the panel location. About 50 feet of extra 3/4 inch EMT was needed.'
        );
      }, 3000);
    }
  };

  const handleCleanup = async () => {
    setIsCleaningUp(true);
    // Simulate AI cleanup
    setTimeout(() => {
      setIsCleaningUp(false);
    }, 2000);
  };

  const handlePhotoUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files) return;

    Array.from(files).forEach((file) => {
      const reader = new FileReader();
      reader.onloadend = () => {
        setPhotos((prev) => [
          ...prev,
          { id: generateId(), url: reader.result as string, caption: '' },
        ]);
      };
      reader.readAsDataURL(file);
    });
  };

  const removePhoto = (photoId: string) => {
    setPhotos((prev) => prev.filter((p) => p.id !== photoId));
  };

  const onSubmit = (data: TicketFormData, status: 'draft' | 'pending_review' = 'draft') => {
    const ticketData: Ticket = {
      id: existingTicket?.id || generateId(),
      companyId: 'comp-001',
      projectId: data.projectId,
      project: projects.find((p) => p.id === data.projectId),
      createdById: 'user-001',
      ticketNumber: existingTicket?.ticketNumber || `TM-2024-${String(tickets.length + 1).padStart(4, '0')}`,
      ticketType: 'time_materials',
      title: data.title,
      descriptionRaw: data.descriptionRaw,
      workDate: data.workDate,
      status,
      lineItems: data.lineItems.map((li, idx) => ({
        id: generateId(),
        ticketId: existingTicket?.id || '',
        itemType: li.itemType,
        description: li.description,
        quantity: li.quantity,
        unit: li.unit,
        unitPrice: li.unitPrice,
        total: li.quantity * li.unitPrice,
        sortOrder: idx + 1,
      })),
      photos: photos.map((p) => ({
        id: p.id,
        ticketId: existingTicket?.id || '',
        fileUrl: p.url,
        thumbnailUrl: p.url,
        caption: p.caption,
        takenAt: new Date().toISOString(),
      })),
      signatures: existingTicket?.signatures || [],
      total,
      createdAt: existingTicket?.createdAt || new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    if (isEdit) {
      updateTicket(existingTicket!.id, ticketData);
    } else {
      addTicket(ticketData);
    }

    navigate('/tickets');
  };

  return (
    <motion.div variants={containerVariants} initial="hidden" animate="visible" className="space-y-6 max-w-4xl mx-auto">
      {/* Header */}
      <motion.div variants={itemVariants} className="flex items-center gap-4">
        <Button variant="ghost" size="default" onClick={() => navigate(-1)} leftIcon={<ArrowLeft className="w-5 h-5" />}>
          Back
        </Button>
        <div>
          <h1 className="text-3xl">{isEdit ? 'Edit Ticket' : 'New T&M Ticket'}</h1>
          {existingTicket && (
            <p className="text-secondary">{existingTicket.ticketNumber}</p>
          )}
        </div>
      </motion.div>

      <form onSubmit={handleSubmit((data) => onSubmit(data, 'draft'))}>
        {/* Basic Info */}
        <motion.div variants={itemVariants}>
          <Card variant="default" padding="lg" className="mb-6">
            <h2 className="text-xl mb-4">Basic Information</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <Select
                label="Project"
                options={projects.map((p) => ({ value: p.id, label: p.name }))}
                error={errors.projectId?.message}
                {...register('projectId', { required: 'Project is required' })}
              />
              <Input
                label="Work Date"
                type="date"
                error={errors.workDate?.message}
                {...register('workDate', { required: 'Work date is required' })}
              />
            </div>
            <div className="mt-4">
              <Input
                label="Title"
                placeholder="Brief description of the extra work"
                error={errors.title?.message}
                {...register('title', { required: 'Title is required' })}
              />
            </div>
          </Card>
        </motion.div>

        {/* Description */}
        <motion.div variants={itemVariants}>
          <Card variant="default" padding="lg" className="mb-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl">Description</h2>
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant={isRecording ? 'destructive' : 'outline'}
                  size="sm"
                  onClick={handleVoiceRecord}
                  leftIcon={isRecording ? <MicOff className="w-4 h-4" /> : <Mic className="w-4 h-4" />}
                >
                  {isRecording ? 'Stop' : 'Voice'}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleCleanup}
                  isLoading={isCleaningUp}
                  leftIcon={<Sparkles className="w-4 h-4" />}
                >
                  Clean Up
                </Button>
              </div>
            </div>
            {isRecording && (
              <div className="mb-4 p-4 bg-red-50 dark:bg-red-900/20 rounded-lg flex items-center gap-3">
                <div className="w-3 h-3 rounded-full bg-red-500 animate-pulse" />
                <span className="text-red-700 dark:text-red-400">Recording... Speak clearly.</span>
              </div>
            )}
            <Textarea
              placeholder="Describe the extra work performed. You can use voice recording for hands-free entry."
              {...register('descriptionRaw')}
              className="min-h-[120px]"
            />
          </Card>
        </motion.div>

        {/* Line Items */}
        <motion.div variants={itemVariants}>
          <Card variant="default" padding="lg" className="mb-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl">Line Items</h2>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => append({ itemType: 'labor', description: '', quantity: 1, unit: 'hours', unitPrice: 85 })}
                leftIcon={<Plus className="w-4 h-4" />}
              >
                Add Item
              </Button>
            </div>

            <div className="space-y-4">
              {fields.map((field, index) => (
                <div key={field.id} className="p-4 bg-charcoal-50 dark:bg-charcoal-800/50 rounded-lg">
                  <div className="flex items-start gap-4 mb-3">
                    <Badge variant={watchLineItems[index]?.itemType === 'labor' ? 'info' : watchLineItems[index]?.itemType === 'material' ? 'success' : 'warning'}>
                      {watchLineItems[index]?.itemType}
                    </Badge>
                    <div className="flex-1">
                      <select
                        className="w-full p-2 rounded border border-default bg-secondary text-sm"
                        {...register(`lineItems.${index}.itemType` as const)}
                      >
                        {itemTypeOptions.map((opt) => (
                          <option key={opt.value} value={opt.value}>{opt.label}</option>
                        ))}
                      </select>
                    </div>
                    <Button type="button" variant="ghost" size="sm" onClick={() => remove(index)}>
                      <Trash2 className="w-4 h-4 text-red-500" />
                    </Button>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-12 gap-3">
                    <div className="sm:col-span-5">
                      <Input
                        placeholder="Description"
                        {...register(`lineItems.${index}.description` as const)}
                      />
                    </div>
                    <div className="sm:col-span-2">
                      <Input
                        type="number"
                        step="0.01"
                        placeholder="Qty"
                        {...register(`lineItems.${index}.quantity` as const, { valueAsNumber: true })}
                      />
                    </div>
                    <div className="sm:col-span-2">
                      <select
                        className="w-full h-12 px-3 rounded-lg border-2 border-default bg-secondary"
                        {...register(`lineItems.${index}.unit` as const)}
                      >
                        {unitOptions.map((opt) => (
                          <option key={opt.value} value={opt.value}>{opt.label}</option>
                        ))}
                      </select>
                    </div>
                    <div className="sm:col-span-3">
                      <Input
                        type="number"
                        step="0.01"
                        placeholder="Unit Price"
                        leftIcon={<span className="text-charcoal-400">$</span>}
                        {...register(`lineItems.${index}.unitPrice` as const, { valueAsNumber: true })}
                      />
                    </div>
                  </div>

                  <div className="text-right mt-2">
                    <span className="text-sm text-secondary">Subtotal: </span>
                    <span className="font-medium">
                      {formatCurrency((watchLineItems[index]?.quantity || 0) * (watchLineItems[index]?.unitPrice || 0))}
                    </span>
                  </div>
                </div>
              ))}
            </div>

            {/* Total */}
            <div className="mt-6 pt-4 border-t border-default flex items-center justify-between">
              <span className="text-lg font-medium">Total</span>
              <span className="font-display text-3xl text-safety-500">{formatCurrency(total)}</span>
            </div>
          </Card>
        </motion.div>

        {/* Photos */}
        <motion.div variants={itemVariants}>
          <Card variant="default" padding="lg" className="mb-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl">Photos</h2>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => fileInputRef.current?.click()}
                leftIcon={<Camera className="w-4 h-4" />}
              >
                Add Photo
              </Button>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                multiple
                className="hidden"
                onChange={handlePhotoUpload}
              />
            </div>

            {photos.length > 0 ? (
              <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
                {photos.map((photo) => (
                  <div key={photo.id} className="relative group">
                    <div className="aspect-square rounded-lg overflow-hidden bg-charcoal-100 dark:bg-charcoal-800">
                      <img src={photo.url} alt="" className="w-full h-full object-cover" />
                    </div>
                    <button
                      type="button"
                      onClick={() => removePhoto(photo.id)}
                      className="absolute top-2 right-2 p-1.5 rounded-full bg-charcoal-900/70 text-white opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <div className="p-8 border-2 border-dashed border-charcoal-200 dark:border-charcoal-700 rounded-lg text-center">
                <ImageIcon className="w-12 h-12 mx-auto text-charcoal-300 dark:text-charcoal-600 mb-3" />
                <p className="text-secondary mb-2">No photos added</p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => fileInputRef.current?.click()}
                >
                  Upload Photos
                </Button>
              </div>
            )}
          </Card>
        </motion.div>

        {/* Actions */}
        <motion.div variants={itemVariants} className="flex flex-col sm:flex-row gap-3">
          <Button type="submit" variant="secondary" size="lg" leftIcon={<Save className="w-5 h-5" />} className="flex-1">
            Save as Draft
          </Button>
          <Button
            type="button"
            onClick={handleSubmit((data) => onSubmit(data, 'pending_review'))}
            size="lg"
            leftIcon={<Send className="w-5 h-5" />}
            className="flex-1"
          >
            Submit for Review
          </Button>
        </motion.div>
      </form>
    </motion.div>
  );
}
