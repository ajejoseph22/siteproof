import { useRef, useState } from 'react';
import SignatureCanvas from 'react-signature-canvas';
import { motion } from 'framer-motion';
import { Eraser, Check, X, Smartphone, Send, PenTool, CheckCircle2 } from 'lucide-react';
import { Button } from '@/components/ui';
import type { Project } from '@/types';

type SignatureStep = 'foreman' | 'gc_send' | 'gc_verify' | 'gc_sign' | 'complete';

interface SignatureCaptureProps {
  project: Project;
  foremanName: string;
  onComplete: (signatures: {
    foremanSignature: string;
    gcSignature?: string;
    gcSignerName?: string;
    signingMethod: 'link_own_device' | 'same_device_verified';
  }) => void;
  onCancel: () => void;
}

export function SignatureCapture({ project, foremanName, onComplete, onCancel }: SignatureCaptureProps) {
  const sigCanvas = useRef<SignatureCanvas | null>(null);
  const [step, setStep] = useState<SignatureStep>('foreman');
  const [foremanSignature, setForemanSignature] = useState<string | null>(null);
  const [verificationCode, setVerificationCode] = useState('');
  const [isSending, setIsSending] = useState(false);
  const [gcSignerName, setGcSignerName] = useState(project.gcContactName || '');

  const handleClear = () => {
    sigCanvas.current?.clear();
  };

  const handleForemanSign = () => {
    if (sigCanvas.current && !sigCanvas.current.isEmpty()) {
      const dataUrl = sigCanvas.current.toDataURL('image/png');
      setForemanSignature(dataUrl);
      sigCanvas.current.clear();
      setStep('gc_send');
    }
  };

  const handleSendToGC = async () => {
    setIsSending(true);
    // In real app, this calls POST /api/v1/tickets/{id}/signature-request
    // with project.gcContactEmail and project.gcContactPhone
    await new Promise((resolve) => setTimeout(resolve, 1500));
    setIsSending(false);
    // For demo, we allow same-device signing with verification
    setStep('gc_verify');
  };

  const handleVerify = async () => {
    // In real app, this calls POST /api/v1/signing-links/{token}/verify
    await new Promise((resolve) => setTimeout(resolve, 500));
    setStep('gc_sign');
  };

  const handleGCSign = () => {
    if (sigCanvas.current && !sigCanvas.current.isEmpty()) {
      const gcSigData = sigCanvas.current.toDataURL('image/png');
      onComplete({
        foremanSignature: foremanSignature!,
        gcSignature: gcSigData,
        gcSignerName: gcSignerName,
        signingMethod: 'same_device_verified',
      });
    }
  };

  const handleSkipGCSignature = () => {
    // Foreman submits, GC will sign later via link
    onComplete({
      foremanSignature: foremanSignature!,
      signingMethod: 'link_own_device',
    });
  };

  // ─────────────────────────────────────────────────────────────────────────────
  // Step 1: Foreman Signs
  // ─────────────────────────────────────────────────────────────────────────────
  if (step === 'foreman') {
    return (
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="space-y-6"
      >
        <div className="text-center mb-4">
          <div className="w-16 h-16 rounded-full bg-safety-500/10 flex items-center justify-center mx-auto mb-4">
            <PenTool className="w-8 h-8 text-safety-500" />
          </div>
          <h3 className="font-display text-xl mb-2">FOREMAN SIGNATURE</h3>
          <p className="text-charcoal-500 dark:text-charcoal-400 text-sm">
            {foremanName}, sign below to certify this work was performed.
          </p>
        </div>

        {/* Signature Pad */}
        <div className="relative">
          <div className="border-2 border-charcoal-300 dark:border-charcoal-600 rounded-xl overflow-hidden bg-white">
            <SignatureCanvas
              ref={sigCanvas}
              canvasProps={{
                className: 'w-full h-48',
                style: { width: '100%', height: '192px' },
              }}
              penColor="#111113"
              backgroundColor="white"
            />
          </div>
          <div className="absolute bottom-8 left-8 right-8 border-b border-charcoal-300" />
          <p className="absolute bottom-2 left-8 text-xs text-charcoal-400">Sign above the line</p>
          <button
            type="button"
            onClick={handleClear}
            className="absolute top-2 right-2 p-2 rounded-lg bg-charcoal-100 hover:bg-charcoal-200 text-charcoal-600 transition-colors"
          >
            <Eraser className="w-4 h-4" />
          </button>
        </div>

        <div className="flex gap-3">
          <Button variant="outline" onClick={onCancel} leftIcon={<X className="w-4 h-4" />} className="flex-1">
            Cancel
          </Button>
          <Button onClick={handleForemanSign} leftIcon={<Check className="w-4 h-4" />} className="flex-1">
            Sign & Continue
          </Button>
        </div>
      </motion.div>
    );
  }

  // ─────────────────────────────────────────────────────────────────────────────
  // Step 2: Send to GC for Signature
  // ─────────────────────────────────────────────────────────────────────────────
  if (step === 'gc_send') {
    return (
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="space-y-6"
      >
        <div className="text-center mb-4">
          <div className="w-16 h-16 rounded-full bg-emerald-500/10 flex items-center justify-center mx-auto mb-4">
            <CheckCircle2 className="w-8 h-8 text-emerald-500" />
          </div>
          <h3 className="font-display text-xl mb-2">YOUR SIGNATURE CAPTURED</h3>
          <p className="text-charcoal-500 dark:text-charcoal-400 text-sm">
            Now get the GC superintendent's signature.
          </p>
        </div>

        {/* GC Info from Project */}
        <div className="p-4 rounded-xl bg-charcoal-50 dark:bg-charcoal-800 border-2 border-charcoal-200 dark:border-charcoal-700">
          <h4 className="font-semibold text-charcoal-900 dark:text-white mb-3">GC Contact</h4>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-charcoal-500 dark:text-charcoal-400">Company:</span>
              <span className="font-medium text-charcoal-900 dark:text-white">{project.gcCompanyName}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-charcoal-500 dark:text-charcoal-400">Contact:</span>
              <span className="font-medium text-charcoal-900 dark:text-white">{project.gcContactName}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-charcoal-500 dark:text-charcoal-400">Email:</span>
              <span className="font-medium text-charcoal-900 dark:text-white">{project.gcContactEmail}</span>
            </div>
            {project.gcContactPhone && (
              <div className="flex justify-between">
                <span className="text-charcoal-500 dark:text-charcoal-400">Phone:</span>
                <span className="font-medium text-charcoal-900 dark:text-white">{project.gcContactPhone}</span>
              </div>
            )}
          </div>
        </div>

        <div className="space-y-3">
          <Button
            onClick={handleSendToGC}
            isLoading={isSending}
            leftIcon={<Send className="w-4 h-4" />}
            className="w-full"
            size="lg"
          >
            Send Signing Link to GC
          </Button>

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-charcoal-200 dark:border-charcoal-700" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-3 bg-white dark:bg-charcoal-900 text-charcoal-500">
                or if GC is present
              </span>
            </div>
          </div>

          <Button
            variant="outline"
            onClick={() => setStep('gc_verify')}
            leftIcon={<Smartphone className="w-4 h-4" />}
            className="w-full"
          >
            Sign on This Device (SMS Verify)
          </Button>
        </div>

        <Button variant="ghost" onClick={handleSkipGCSignature} className="w-full text-charcoal-500">
          Submit Without GC Signature (GC signs later via link)
        </Button>
      </motion.div>
    );
  }

  // ─────────────────────────────────────────────────────────────────────────────
  // Step 3: SMS Verification for Same-Device Signing
  // ─────────────────────────────────────────────────────────────────────────────
  if (step === 'gc_verify') {
    return (
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="space-y-6"
      >
        <div className="text-center mb-4">
          <div className="w-16 h-16 rounded-full bg-blue-500/10 flex items-center justify-center mx-auto mb-4">
            <Smartphone className="w-8 h-8 text-blue-500" />
          </div>
          <h3 className="font-display text-xl mb-2">VERIFY GC IDENTITY</h3>
          <p className="text-charcoal-500 dark:text-charcoal-400 text-sm">
            A 6-digit code was sent to {project.gcContactPhone || project.gcContactEmail}
          </p>
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-charcoal-700 dark:text-charcoal-300 mb-2">
              Verification Code
            </label>
            <input
              type="text"
              placeholder="000000"
              value={verificationCode}
              onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
              className="w-full h-14 text-center text-2xl tracking-[0.5em] font-mono rounded-xl border-2 border-charcoal-300 dark:border-charcoal-600 bg-white dark:bg-charcoal-800 focus:border-safety-500 focus:ring-2 focus:ring-safety-500/20 outline-none"
              maxLength={6}
            />
          </div>
          <p className="text-center text-sm text-charcoal-500">
            Didn't receive the code?{' '}
            <button type="button" className="text-safety-500 hover:underline font-medium">
              Resend
            </button>
          </p>
        </div>

        {/* Demo hint */}
        <div className="p-3 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg text-center text-sm">
          <p className="text-amber-700 dark:text-amber-400">
            Demo: Enter any 6 digits to continue
          </p>
        </div>

        <div className="flex gap-3">
          <Button variant="outline" onClick={() => setStep('gc_send')} className="flex-1">
            Back
          </Button>
          <Button
            onClick={handleVerify}
            className="flex-1"
            disabled={verificationCode.length !== 6}
          >
            Verify
          </Button>
        </div>
      </motion.div>
    );
  }

  // ─────────────────────────────────────────────────────────────────────────────
  // Step 4: GC Signs
  // ─────────────────────────────────────────────────────────────────────────────
  if (step === 'gc_sign') {
    return (
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="space-y-6"
      >
        <div className="text-center mb-4">
          <div className="w-16 h-16 rounded-full bg-safety-500/10 flex items-center justify-center mx-auto mb-4">
            <PenTool className="w-8 h-8 text-safety-500" />
          </div>
          <h3 className="font-display text-xl mb-2">GC SUPERINTENDENT SIGNATURE</h3>
          <p className="text-charcoal-500 dark:text-charcoal-400 text-sm">
            {project.gcContactName}, sign below to approve this T&M ticket.
          </p>
        </div>

        {/* Signer Name Confirmation */}
        <div>
          <label className="block text-sm font-medium text-charcoal-700 dark:text-charcoal-300 mb-2">
            Your Full Name
          </label>
          <input
            type="text"
            value={gcSignerName}
            onChange={(e) => setGcSignerName(e.target.value)}
            className="w-full h-12 px-4 rounded-xl border-2 border-charcoal-300 dark:border-charcoal-600 bg-white dark:bg-charcoal-800 focus:border-safety-500 focus:ring-2 focus:ring-safety-500/20 outline-none"
            placeholder="Enter your full name"
          />
        </div>

        {/* Signature Pad */}
        <div className="relative">
          <div className="border-2 border-charcoal-300 dark:border-charcoal-600 rounded-xl overflow-hidden bg-white">
            <SignatureCanvas
              ref={sigCanvas}
              canvasProps={{
                className: 'w-full h-48',
                style: { width: '100%', height: '192px' },
              }}
              penColor="#111113"
              backgroundColor="white"
            />
          </div>
          <div className="absolute bottom-8 left-8 right-8 border-b border-charcoal-300" />
          <p className="absolute bottom-2 left-8 text-xs text-charcoal-400">Sign above the line</p>
          <button
            type="button"
            onClick={handleClear}
            className="absolute top-2 right-2 p-2 rounded-lg bg-charcoal-100 hover:bg-charcoal-200 text-charcoal-600 transition-colors"
          >
            <Eraser className="w-4 h-4" />
          </button>
        </div>

        {/* Legal Acknowledgment */}
        <div className="p-4 rounded-xl bg-charcoal-50 dark:bg-charcoal-800 border border-charcoal-200 dark:border-charcoal-700">
          <label className="flex items-start gap-3 cursor-pointer">
            <input
              type="checkbox"
              defaultChecked
              className="mt-1 w-5 h-5 rounded border-charcoal-300 text-safety-500 focus:ring-safety-500"
            />
            <span className="text-xs text-charcoal-600 dark:text-charcoal-400">
              By signing, I acknowledge that the work described in this ticket was performed as stated
              and I approve the labor, materials, and equipment charges listed. This signature is legally binding.
            </span>
          </label>
        </div>

        <div className="flex gap-3">
          <Button variant="outline" onClick={() => setStep('gc_verify')} leftIcon={<X className="w-4 h-4" />} className="flex-1">
            Back
          </Button>
          <Button
            onClick={handleGCSign}
            leftIcon={<Check className="w-4 h-4" />}
            className="flex-1"
            variant="success"
            disabled={!gcSignerName.trim()}
          >
            Approve & Sign
          </Button>
        </div>
      </motion.div>
    );
  }

  return null;
}
