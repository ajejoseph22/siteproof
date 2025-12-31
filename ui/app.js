/**
 * SiteProof - T&M Documentation App
 * UI Prototype JavaScript
 */

(function() {
    'use strict';

    // ========================================
    // State Management
    // ========================================
    
    const state = {
        currentView: 'dashboard',
        isLoggedIn: false,
        isRecording: false,
        sideNavOpen: false
    };

    // ========================================
    // DOM Elements
    // ========================================
    
    const elements = {
        // Screens
        loginScreen: document.getElementById('login-screen'),
        appShell: document.getElementById('app-shell'),
        
        // Forms
        loginForm: document.getElementById('login-form'),
        
        // Navigation
        menuToggle: document.getElementById('menu-toggle'),
        sideNav: document.getElementById('side-nav'),
        navItems: document.querySelectorAll('.nav-item[data-view]'),
        mobileNavItems: document.querySelectorAll('.mobile-nav-item[data-view]'),
        logoutBtn: document.getElementById('logout-btn'),
        
        // Views
        views: {
            dashboard: document.getElementById('dashboard-view'),
            createTicket: document.getElementById('create-ticket-view'),
            ticketDetail: document.getElementById('ticket-detail-view'),
            signature: document.getElementById('signature-view')
        },
        
        // Buttons
        newTicketBtn: document.getElementById('new-ticket-btn'),
        mobileNewTicket: document.getElementById('mobile-new-ticket'),
        backToDashboard: document.getElementById('back-to-dashboard'),
        backFromDetail: document.getElementById('back-from-detail'),
        voiceInputBtn: document.getElementById('voice-input-btn'),
        addLineItem: document.getElementById('add-line-item'),
        addPhotoBtn: document.getElementById('add-photo-btn'),
        submitForSignature: document.getElementById('submit-for-signature'),
        rejectTicketBtn: document.getElementById('reject-ticket-btn'),
        clearSignature: document.getElementById('clear-signature'),
        approveSignBtn: document.getElementById('approve-sign-btn'),
        
        // Voice
        voiceStatus: document.getElementById('voice-status'),
        
        // Signature
        signaturePad: document.getElementById('signature-pad'),
        
        // Line Items
        lineItemsContainer: document.getElementById('line-items'),
        
        // Toast Container
        toastContainer: document.getElementById('toast-container'),
        
        // Modal
        rejectionModal: document.getElementById('rejection-modal'),
        closeRejectionModal: document.getElementById('close-rejection-modal'),
        cancelRejection: document.getElementById('cancel-rejection'),
        
        // Ticket cards
        ticketCards: document.querySelectorAll('.ticket-card')
    };

    // ========================================
    // Navigation Functions
    // ========================================
    
    function showScreen(screenId) {
        elements.loginScreen.classList.remove('active');
        elements.appShell.classList.remove('active');
        
        if (screenId === 'login') {
            elements.loginScreen.classList.add('active');
        } else if (screenId === 'app') {
            elements.appShell.classList.add('active');
        }
    }
    
    function showView(viewName) {
        // Hide all views
        Object.values(elements.views).forEach(view => {
            if (view) view.classList.remove('active');
        });
        
        // Show requested view
        if (elements.views[viewName]) {
            elements.views[viewName].classList.add('active');
            state.currentView = viewName;
        } else if (viewName === 'dashboard') {
            elements.views.dashboard.classList.add('active');
            state.currentView = 'dashboard';
        }
        
        // Update nav items
        updateNavActiveState(viewName);
        
        // Close side nav on mobile
        closeSideNav();
        
        // Scroll to top
        window.scrollTo(0, 0);
    }
    
    function updateNavActiveState(viewName) {
        // Desktop nav
        elements.navItems.forEach(item => {
            item.classList.toggle('active', item.dataset.view === viewName);
        });
        
        // Mobile nav
        elements.mobileNavItems.forEach(item => {
            item.classList.toggle('active', item.dataset.view === viewName);
        });
    }
    
    function toggleSideNav() {
        state.sideNavOpen = !state.sideNavOpen;
        elements.sideNav.classList.toggle('open', state.sideNavOpen);
    }
    
    function closeSideNav() {
        state.sideNavOpen = false;
        elements.sideNav.classList.remove('open');
    }

    // ========================================
    // Authentication
    // ========================================
    
    function handleLogin(e) {
        e.preventDefault();
        
        // Simulate login
        showToast('Welcome back, Mike!', 'success');
        state.isLoggedIn = true;
        showScreen('app');
        showView('dashboard');
    }
    
    function handleLogout(e) {
        e.preventDefault();
        state.isLoggedIn = false;
        showScreen('login');
        showToast('You have been logged out', 'success');
    }

    // ========================================
    // Voice Input
    // ========================================
    
    function toggleVoiceInput() {
        state.isRecording = !state.isRecording;
        
        if (state.isRecording) {
            elements.voiceInputBtn.classList.add('recording');
            elements.voiceStatus.classList.add('active');
            
            // Simulate voice recording for demo
            setTimeout(() => {
                if (state.isRecording) {
                    stopVoiceInput();
                    simulateTranscription();
                }
            }, 3000);
        } else {
            stopVoiceInput();
        }
    }
    
    function stopVoiceInput() {
        state.isRecording = false;
        elements.voiceInputBtn.classList.remove('recording');
        elements.voiceStatus.classList.remove('active');
    }
    
    function simulateTranscription() {
        const textarea = document.getElementById('ticket-description');
        const sampleText = "Ran 50 linear feet of extra 3/4 inch EMT conduit on floor 3 west wing due to design change per RFI number 47. Work included mounting supports, pulling wire, and making connections at both ends. All work completed to code specifications.";
        
        // Simulate typing effect
        let i = 0;
        textarea.value = '';
        
        function typeChar() {
            if (i < sampleText.length) {
                textarea.value += sampleText.charAt(i);
                i++;
                setTimeout(typeChar, 20);
            } else {
                showToast('Transcription cleaned up', 'success');
            }
        }
        
        typeChar();
    }

    // ========================================
    // Line Items
    // ========================================
    
    function addNewLineItem() {
        const newItem = document.createElement('div');
        newItem.className = 'line-item';
        newItem.innerHTML = `
            <div class="line-item-type">
                <select>
                    <option value="labor">Labor</option>
                    <option value="material">Material</option>
                    <option value="equipment">Equipment</option>
                </select>
            </div>
            <div class="line-item-desc">
                <input type="text" placeholder="Description">
            </div>
            <div class="line-item-qty">
                <input type="number" placeholder="Qty">
            </div>
            <div class="line-item-unit">
                <select>
                    <option value="hours">hours</option>
                    <option value="each">each</option>
                    <option value="linear_ft">linear ft</option>
                    <option value="sq_ft">sq ft</option>
                    <option value="day">day</option>
                </select>
            </div>
            <div class="line-item-rate">
                <input type="number" placeholder="Rate">
            </div>
            <div class="line-item-total">
                <span>$0.00</span>
            </div>
            <button class="line-item-remove">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18"/>
                    <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
            </button>
        `;
        
        elements.lineItemsContainer.appendChild(newItem);
        
        // Add remove handler
        newItem.querySelector('.line-item-remove').addEventListener('click', () => {
            newItem.remove();
            updateTotals();
        });
        
        // Add calculation handlers
        const inputs = newItem.querySelectorAll('input[type="number"]');
        inputs.forEach(input => {
            input.addEventListener('input', () => calculateLineItemTotal(newItem));
        });
    }
    
    function calculateLineItemTotal(lineItem) {
        const qty = parseFloat(lineItem.querySelector('.line-item-qty input').value) || 0;
        const rate = parseFloat(lineItem.querySelector('.line-item-rate input').value) || 0;
        const total = qty * rate;
        lineItem.querySelector('.line-item-total span').textContent = `$${total.toFixed(2)}`;
        updateTotals();
    }
    
    function updateTotals() {
        let subtotal = 0;
        document.querySelectorAll('.line-item').forEach(item => {
            const totalText = item.querySelector('.line-item-total span').textContent;
            const value = parseFloat(totalText.replace('$', '')) || 0;
            subtotal += value;
        });
        
        const summaryRows = document.querySelectorAll('.summary-row .summary-value');
        if (summaryRows.length >= 2) {
            summaryRows[0].textContent = `$${subtotal.toFixed(2)}`;
            summaryRows[1].textContent = `$${subtotal.toFixed(2)}`;
        }
    }

    // ========================================
    // Signature Pad
    // ========================================
    
    let signatureCtx = null;
    let isDrawing = false;
    let lastX = 0;
    let lastY = 0;
    
    function initSignaturePad() {
        const canvas = elements.signaturePad;
        if (!canvas) return;
        
        signatureCtx = canvas.getContext('2d');
        
        // Set canvas size
        resizeCanvas();
        window.addEventListener('resize', resizeCanvas);
        
        // Set drawing style
        signatureCtx.strokeStyle = '#1a1d23';
        signatureCtx.lineWidth = 3;
        signatureCtx.lineCap = 'round';
        signatureCtx.lineJoin = 'round';
        
        // Mouse events
        canvas.addEventListener('mousedown', startDrawing);
        canvas.addEventListener('mousemove', draw);
        canvas.addEventListener('mouseup', stopDrawing);
        canvas.addEventListener('mouseout', stopDrawing);
        
        // Touch events
        canvas.addEventListener('touchstart', handleTouchStart);
        canvas.addEventListener('touchmove', handleTouchMove);
        canvas.addEventListener('touchend', stopDrawing);
    }
    
    function resizeCanvas() {
        const canvas = elements.signaturePad;
        if (!canvas) return;
        
        const rect = canvas.parentElement.getBoundingClientRect();
        canvas.width = rect.width;
        canvas.height = 200;
        
        // Reset style after resize
        if (signatureCtx) {
            signatureCtx.strokeStyle = '#1a1d23';
            signatureCtx.lineWidth = 3;
            signatureCtx.lineCap = 'round';
            signatureCtx.lineJoin = 'round';
        }
    }
    
    function startDrawing(e) {
        isDrawing = true;
        const rect = elements.signaturePad.getBoundingClientRect();
        lastX = e.clientX - rect.left;
        lastY = e.clientY - rect.top;
    }
    
    function draw(e) {
        if (!isDrawing) return;
        
        const rect = elements.signaturePad.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        signatureCtx.beginPath();
        signatureCtx.moveTo(lastX, lastY);
        signatureCtx.lineTo(x, y);
        signatureCtx.stroke();
        
        lastX = x;
        lastY = y;
    }
    
    function handleTouchStart(e) {
        e.preventDefault();
        const touch = e.touches[0];
        const mouseEvent = new MouseEvent('mousedown', {
            clientX: touch.clientX,
            clientY: touch.clientY
        });
        elements.signaturePad.dispatchEvent(mouseEvent);
    }
    
    function handleTouchMove(e) {
        e.preventDefault();
        const touch = e.touches[0];
        const mouseEvent = new MouseEvent('mousemove', {
            clientX: touch.clientX,
            clientY: touch.clientY
        });
        elements.signaturePad.dispatchEvent(mouseEvent);
    }
    
    function stopDrawing() {
        isDrawing = false;
    }
    
    function clearSignaturePad() {
        if (!signatureCtx) return;
        signatureCtx.clearRect(0, 0, elements.signaturePad.width, elements.signaturePad.height);
    }

    // ========================================
    // Toast Notifications
    // ========================================
    
    function showToast(message, type = 'success') {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        
        const iconSvg = type === 'success' 
            ? '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22,4 12,14.01 9,11.01"/></svg>'
            : '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>';
        
        toast.innerHTML = `
            <span class="toast-icon">${iconSvg}</span>
            <span class="toast-message">${message}</span>
        `;
        
        elements.toastContainer.appendChild(toast);
        
        // Auto remove after 4 seconds
        setTimeout(() => {
            toast.style.animation = 'fadeIn 0.3s ease reverse';
            setTimeout(() => toast.remove(), 300);
        }, 4000);
    }

    // ========================================
    // Modal Functions
    // ========================================
    
    function openRejectionModal() {
        elements.rejectionModal.classList.add('open');
    }
    
    function closeRejectionModalFn() {
        elements.rejectionModal.classList.remove('open');
    }

    // ========================================
    // Event Listeners
    // ========================================
    
    function initEventListeners() {
        // Login
        if (elements.loginForm) {
            elements.loginForm.addEventListener('submit', handleLogin);
        }
        
        // Logout
        if (elements.logoutBtn) {
            elements.logoutBtn.addEventListener('click', handleLogout);
        }
        
        // Navigation toggle
        if (elements.menuToggle) {
            elements.menuToggle.addEventListener('click', toggleSideNav);
        }
        
        // Click outside to close nav
        document.addEventListener('click', (e) => {
            if (state.sideNavOpen && 
                !elements.sideNav.contains(e.target) && 
                !elements.menuToggle.contains(e.target)) {
                closeSideNav();
            }
        });
        
        // Nav items
        elements.navItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const view = item.dataset.view;
                if (view) showView(view);
            });
        });
        
        // Mobile nav items
        elements.mobileNavItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const view = item.dataset.view;
                if (view) showView(view);
            });
        });
        
        // New ticket buttons
        if (elements.newTicketBtn) {
            elements.newTicketBtn.addEventListener('click', () => showView('createTicket'));
        }
        if (elements.mobileNewTicket) {
            elements.mobileNewTicket.addEventListener('click', () => showView('createTicket'));
        }
        
        // Back buttons
        if (elements.backToDashboard) {
            elements.backToDashboard.addEventListener('click', () => showView('dashboard'));
        }
        if (elements.backFromDetail) {
            elements.backFromDetail.addEventListener('click', () => showView('dashboard'));
        }
        
        // Voice input
        if (elements.voiceInputBtn) {
            elements.voiceInputBtn.addEventListener('click', toggleVoiceInput);
        }
        
        // Add line item
        if (elements.addLineItem) {
            elements.addLineItem.addEventListener('click', addNewLineItem);
        }
        
        // Add photo
        if (elements.addPhotoBtn) {
            elements.addPhotoBtn.addEventListener('click', () => {
                showToast('Camera would open here', 'success');
            });
        }
        
        // Submit for signature
        if (elements.submitForSignature) {
            elements.submitForSignature.addEventListener('click', () => {
                showToast('Signature request sent!', 'success');
                showView('dashboard');
            });
        }
        
        // Ticket cards
        elements.ticketCards.forEach(card => {
            card.addEventListener('click', () => {
                showView('ticketDetail');
            });
        });
        
        // Clear signature
        if (elements.clearSignature) {
            elements.clearSignature.addEventListener('click', clearSignaturePad);
        }
        
        // Reject ticket
        if (elements.rejectTicketBtn) {
            elements.rejectTicketBtn.addEventListener('click', openRejectionModal);
        }
        
        // Modal close buttons
        if (elements.closeRejectionModal) {
            elements.closeRejectionModal.addEventListener('click', closeRejectionModalFn);
        }
        if (elements.cancelRejection) {
            elements.cancelRejection.addEventListener('click', closeRejectionModalFn);
        }
        
        // Modal backdrop click
        document.querySelectorAll('.modal-backdrop').forEach(backdrop => {
            backdrop.addEventListener('click', () => {
                closeRejectionModalFn();
            });
        });
        
        // Approve & Sign
        if (elements.approveSignBtn) {
            elements.approveSignBtn.addEventListener('click', () => {
                const checkbox = document.getElementById('acknowledge-checkbox');
                const signerName = document.getElementById('signer-name');
                
                if (!checkbox.checked) {
                    showToast('Please acknowledge the terms', 'error');
                    return;
                }
                
                if (!signerName.value.trim()) {
                    showToast('Please enter your name', 'error');
                    return;
                }
                
                showToast('Ticket approved and signed!', 'success');
                showView('dashboard');
            });
        }
        
        // Line item calculations
        document.querySelectorAll('.line-item').forEach(item => {
            const inputs = item.querySelectorAll('input[type="number"]');
            inputs.forEach(input => {
                input.addEventListener('input', () => calculateLineItemTotal(item));
            });
            
            const removeBtn = item.querySelector('.line-item-remove');
            if (removeBtn) {
                removeBtn.addEventListener('click', () => {
                    item.remove();
                    updateTotals();
                });
            }
        });
        
        // Toggle buttons for ticket type
        document.querySelectorAll('.toggle-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                document.querySelectorAll('.toggle-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
            });
        });
        
        // View full details button (signature page)
        document.querySelectorAll('.view-full-btn').forEach(btn => {
            btn.addEventListener('click', () => showView('ticketDetail'));
        });
        
        // Link buttons with data-view
        document.querySelectorAll('[data-view]').forEach(el => {
            if (!el.classList.contains('nav-item') && !el.classList.contains('mobile-nav-item')) {
                el.addEventListener('click', (e) => {
                    e.preventDefault();
                    const view = el.dataset.view;
                    if (view) showView(view);
                });
            }
        });
    }

    // ========================================
    // Initialize App
    // ========================================
    
    function init() {
        initEventListeners();
        initSignaturePad();
        
        // Show login screen by default
        showScreen('login');
        
        // For demo: auto-login after 1 second if hash is #app
        if (window.location.hash === '#app') {
            setTimeout(() => {
                state.isLoggedIn = true;
                showScreen('app');
                showView('dashboard');
            }, 500);
        }
        
        // For demo: show signature view if hash is #sign
        if (window.location.hash === '#sign') {
            setTimeout(() => {
                state.isLoggedIn = true;
                showScreen('app');
                showView('signature');
            }, 500);
        }
    }
    
    // Run on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();

