// CTF Challenge App - Simple Frontend
class CTFApp {
    constructor() {
        this.challenges = [];
        this.currentChallenge = null;
        this.userInfo = null;
        this.challengeAccessed = {}; // Track accessed/viewed challenges
        
        // Load accessed challenges from localStorage
        this.loadAccessedChallenges();
        
        // Set up auto-refresh for token expiration
        this.setupAutoRefresh();
        
        this.init();
    }

    // Cache user info to localStorage
    cacheUserInfo(userInfo) {
        try {
            localStorage.setItem('ctf_user_cache', JSON.stringify({
                username: userInfo.user_email || userInfo.username,
                cached_at: Date.now()
            }));
        } catch (error) {
            console.warn('Failed to cache user info:', error);
        }
    }

    // Load accessed challenges from localStorage
    loadAccessedChallenges() {
        try {
            const accessed = localStorage.getItem('ctf_accessed_challenges');
            if (accessed) {
                this.challengeAccessed = JSON.parse(accessed);
            }
        } catch (error) {
            console.warn('Failed to load accessed challenges:', error);
        }
    }

    // Save accessed challenges to localStorage
    saveAccessedChallenges() {
        try {
            localStorage.setItem('ctf_accessed_challenges', JSON.stringify(this.challengeAccessed));
        } catch (error) {
            console.warn('Failed to save accessed challenges:', error);
        }
    }

    // Get cached user info from localStorage
    getCachedUserInfo() {
        try {
            const cached = localStorage.getItem('ctf_user_cache');
            if (cached) {
                const data = JSON.parse(cached);
                // Use cache if it's less than 24 hours old
                if (Date.now() - data.cached_at < 24 * 60 * 60 * 1000) {
                    return { user_email: data.username, tokens: 0, points: 0 };
                }
            }
        } catch (error) {
            console.warn('Failed to get cached user info:', error);
        }
        return null;
    }

    init() {
        this.checkAuth();
        this.bindEvents();
        this.handleRouting();
    }

    // Handle URL routing to maintain state on refresh
    handleRouting() {
        const urlParams = new URLSearchParams(window.location.search);
        const challengeId = urlParams.get('challenge');
        
        if (challengeId) {
            // Store the challenge ID to load after authentication
            this.pendingChallengeId = challengeId;
            // Pre-hide the challenges list to avoid flash
            document.getElementById('challengesList').style.display = 'none';
            document.getElementById('infoSection').style.display = 'none';
            document.getElementById('challengeDetail').style.display = 'block';
        }
    }


    // Pretty print category names
    prettifyCategory(category) {
        return category
            .split('-')
            .map(word => word.charAt(0).toUpperCase() + word.slice(1))
            .join(' ');
    }

    // Get AWS Verified Access JWT from cookie (for debugging only)
    getAuthToken() {
        const cookies = document.cookie.split(';');
        for (let cookie of cookies) {
            const idx = cookie.trim().indexOf('=');
            const name = cookie.trim().substring(0, idx);
            const value = cookie.trim().substring(idx + 1);
            if (name === 'AWSVAAuthSessionCookie') {
                return value;
            }
        }
        return null;
    }

    // Check authentication and load user info
    async checkAuth() {
        try {
            // Try to load challenges directly - backend will handle auth via cookies
            await this.loadChallenges();
            
            // Load user profile after successful auth test
            await this.loadUserProfile();
            this.showUserInfo();
            
            // Check if we need to show a specific challenge from URL
            if (this.pendingChallengeId) {
                const challenge = this.findChallengeById(this.pendingChallengeId);
                if (challenge) {
                    this.showChallenge(challenge);
                } else {
                    // Invalid challenge ID, redirect to homepage
                    this.updateURL();
                    this.showChallengesList();
                }
                this.pendingChallengeId = null;
            }
        } catch (error) {
            this.showAuthError();
        }
    }

    // Find challenge by ID
    findChallengeById(challengeId) {
        for (const category of this.challenges) {
            for (const challenge of category.challenges) {
                if (challenge.id.toString() === challengeId) {
                    return challenge;
                }
            }
        }
        return null;
    }

    // Update challenge status in the challenges array
    updateChallengeStatus(challengeId, statusType, value) {
        for (const category of this.challenges) {
            for (const challenge of category.challenges) {
                if (challenge.id.toString() === challengeId.toString()) {
                    challenge[statusType] = value;
                    return;
                }
            }
        }
    }

    // Load user profile from API
    async loadUserProfile() {
        try {
            const response = await this.apiCall('/api/profile');
            if (response.ok) {
                this.userInfo = await response.json();
                this.cacheUserInfo(this.userInfo);
            } else {
                // Try to use cached user info
                const cachedUser = this.getCachedUserInfo();
                this.userInfo = cachedUser || {
                    user_email: 'user_email',
                    tokens: 0,
                    points: 0
                };
            }
        } catch (error) {
            // Try to use cached user info
            const cachedUser = this.getCachedUserInfo();
            this.userInfo = cachedUser || {
                user_email: 'user_email',
                tokens: 0,
                points: 0
            };
        }
    }


    // Make API call with authentication
    async apiCall(endpoint, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json'
            },
            credentials: 'include' // Include cookies for authentication
        };

        const mergedOptions = {
            ...defaultOptions,
            ...options,
            headers: {
                ...defaultOptions.headers,
                ...(options.headers || {})
            }
        };

        try {
            const response = await fetch(endpoint, mergedOptions);
            
            // Handle 401 (token expired) by refreshing the page
            if (response.status === 401) {
                window.location.reload();
                return;
            }
            
            return response;
        } catch (error) {
            this.showError(`Network error: ${error.message}`);
            throw error;
        }
    }

    // Set up auto-refresh to handle token expiration
    setupAutoRefresh() {
        // Refresh every 30 minutes to prevent token expiration
        setInterval(() => {
            window.location.reload();
        }, 30 * 60 * 1000); // 30 minutes
    }

    // Show authentication error
    showAuthError() {
        document.getElementById('authError').style.display = 'block';
        document.getElementById('challengesList').style.display = 'none';
        document.getElementById('infoSection').style.display = 'none';
        document.getElementById('flagNote').style.display = 'none';
        document.getElementById('challengeDetail').style.display = 'none';
    }

    // Show challenges list
    showChallengesList() {
        document.getElementById('challengesList').style.display = 'block';
        document.getElementById('infoSection').style.display = 'block';
        document.getElementById('flagNote').style.display = 'block';
        document.getElementById('examSection').style.display = 'block';
        document.getElementById('examList').style.display = 'none';
        document.getElementById('challengeDetail').style.display = 'none';
        document.getElementById('authError').style.display = 'none';
    }

    // Show user info in header
    showUserInfo() {
        if (this.userInfo) {
            const username = this.userInfo.user_email || this.userInfo.username || 'unknownusername';
            document.getElementById('username').textContent = username;
            document.getElementById('userTokens').textContent = `${this.userInfo.tokens || 0} tokens`;
            document.getElementById('userPoints').textContent = `${this.userInfo.points || 0} points`;
            
            // Show alias if available
            if (this.userInfo.alias) {
                document.getElementById('userAliasText').textContent = this.userInfo.alias;
                document.getElementById('userAlias').style.display = 'inline';
            } else {
                document.getElementById('userAlias').style.display = 'none';
            }
            
            document.getElementById('userInfo').style.display = 'flex';
            
            // Only show exam button when on the main challenges page
            const isOnMainPage = document.getElementById('challengesList').style.display === 'block';
            if (isOnMainPage) {
                document.getElementById('examSection').style.display = 'block';
            }
        }
    }

    // Load challenges from API
    async loadChallenges() {
        try {
            const response = await this.apiCall('/api/challenges');
            if (response.ok) {
                const challengesList = await response.json();
                this.challenges = this.groupChallengesByCategory(challengesList);
                this.renderChallenges();
            } else {
                const errorData = await response.json().catch(() => ({}));
                const errorMessage = errorData.error || errorData.message || `HTTP ${response.status}: Failed to load challenges`;
                throw new Error(errorMessage);
            }
        } catch (error) {
            throw error; // Re-throw for checkAuth to handle
        }
    }

    // Group challenges by category
    groupChallengesByCategory(challengesList) {
        const grouped = {};
        challengesList.forEach(challenge => {
            if (!grouped[challenge.category]) {
                grouped[challenge.category] = {
                    category: challenge.category,
                    challenges: []
                };
            }
            
            grouped[challenge.category].challenges.push({
                id: challenge.id,
                nested_id: challenge.nested_id,
                name: challenge.name,
                description: challenge.description,
                point_reward_amount: challenge.point_reward_amount,
                category: challenge.category,
                solved: challenge.completed || false,
                accessed: this.challengeAccessed[challenge.id] || false
            });
        });
        
        // Sort challenges within each category by nested_id
        Object.values(grouped).forEach(category => {
            category.challenges.sort((a, b) => a.nested_id - b.nested_id);
        });
        
        return Object.values(grouped);
    }


    // Render challenges list
    renderChallenges() {
        const container = document.getElementById('challengesList');
        container.innerHTML = '';
        
        // Only show challenges list if we're not loading a specific challenge
        if (!this.pendingChallengeId) {
            this.showChallengesList();
        }

        this.challenges.forEach(category => {
            const categoryDiv = document.createElement('div');
            categoryDiv.className = 'challenge-category';
            
            const categoryTitle = document.createElement('h2');
            categoryTitle.textContent = this.prettifyCategory(category.category);
            categoryDiv.appendChild(categoryTitle);
            
            const gridDiv = document.createElement('div');
            gridDiv.className = 'challenge-grid';
            
            category.challenges.forEach(challenge => {
                const cardDiv = document.createElement('div');
                cardDiv.className = 'challenge-card';
                
                // Add status classes - solved takes precedence over accessed
                if (challenge.solved) {
                    cardDiv.classList.add('completed');
                } else if (challenge.accessed) {
                    cardDiv.classList.add('accessed');
                }
                
                const statusText = challenge.solved ? '★ Solved' : 
                                  (challenge.accessed ? '◐ Viewed' : '');
                const statusClass = challenge.solved ? 'status-solved' : 
                                   (challenge.accessed ? 'status-accessed' : '');
                
                cardDiv.innerHTML = `
                    <div class="challenge-name">${challenge.name}</div>
                    <div class="challenge-description">${challenge.description}</div>
                    <div class="challenge-meta">
                        <span class="challenge-points">${challenge.point_reward_amount} ${challenge.point_reward_amount === 1 ? 'point' : 'points'}</span>
                        ${statusText ? `<span class="challenge-status ${statusClass}">${statusText}</span>` : ''}
                    </div>
                `;
                
                cardDiv.addEventListener('click', () => this.showChallenge(challenge));
                gridDiv.appendChild(cardDiv);
            });
            
            categoryDiv.appendChild(gridDiv);
            container.appendChild(categoryDiv);
        });
    }

    // Show individual challenge detail
    async showChallenge(challenge) {
        this.currentChallenge = challenge;
        
        // Mark challenge as accessed
        this.challengeAccessed[challenge.id] = true;
        this.saveAccessedChallenges();
        
        // Update the challenge in the challenges array to reflect accessed status
        this.updateChallengeStatus(challenge.id, 'accessed', true);
        
        // Update URL to include challenge ID
        this.updateURL(challenge.id);
        
        // Clear previous challenge information immediately to prevent flashing
        this.clearChallengeDetails();
        
        document.getElementById('challengesList').style.display = 'none';
        document.getElementById('infoSection').style.display = 'none';
        document.getElementById('flagNote').style.display = 'none';
        document.getElementById('examSection').style.display = 'none';
        document.getElementById('examList').style.display = 'none';
        document.getElementById('challengeDetail').style.display = 'block';
        
        // Load detailed challenge info
        await this.loadChallengeDetails(challenge.id);
    }

    // Clear challenge details to prevent information from previous challenges lingering
    clearChallengeDetails() {
        document.getElementById('challengeTitle').textContent = 'Loading...';
        document.getElementById('challengeDescription').textContent = '';
        document.getElementById('challengePoints').textContent = '';
        document.getElementById('challengeStatus').textContent = '';
        document.getElementById('challengeStatus').className = 'status';
        document.getElementById('challengeTextAsset').innerHTML = '';
        document.getElementById('challengeFiles').innerHTML = '';
        
        // Clear previous flag result
        const flagResult = document.getElementById('flagResult');
        flagResult.innerHTML = '';
        flagResult.className = 'flag-result';
        flagResult.style.display = 'none';
        document.getElementById('flagInput').value = '';
    }
    
    // Load detailed challenge information
    async loadChallengeDetails(challengeId) {
        try {
            const response = await this.apiCall(`/api/challenges/${challengeId}`);
            if (response.ok) {
                const detailedChallenge = await response.json();
                this.renderChallengeDetails(detailedChallenge);
            } else {
                const errorData = await response.json().catch(() => ({}));
                const errorMessage = errorData.error || errorData.message || `HTTP ${response.status}: Failed to load challenge details`;
                throw new Error(errorMessage);
            }
        } catch (error) {
            console.error('Error loading challenge details:', error);
            this.showError(`Failed to load challenge details: ${error.message}`);
        }
    }
    
    // Render challenge details
    renderChallengeDetails(challenge) {
        document.getElementById('challengeTitle').textContent = challenge.name;
        document.getElementById('challengeDescription').textContent = challenge.description;
        
        // Show points or token cost based on challenge type
        if (challenge.is_exam) {
            document.getElementById('challengePoints').textContent = '1 token cost';
        } else {
            document.getElementById('challengePoints').textContent = `${challenge.point_reward_amount} ${challenge.point_reward_amount === 1 ? 'point' : 'points'}`;
        }
        
        // Update current challenge reference
        this.currentChallenge = {
            ...this.currentChallenge,
            ...challenge
        };
        
        // Show status - check if solved
        const statusElement = document.getElementById('challengeStatus');
        const solved = challenge.completed || false;
        
        if (solved) {
            statusElement.textContent = 'Solved ★';
            statusElement.className = 'status completed';
        } else {
            statusElement.textContent = 'Not completed';
            statusElement.className = 'status';
        }
        
        // Show text asset if available
        const textAssetContainer = document.getElementById('challengeTextAsset');
        if (challenge.text_asset) {
            textAssetContainer.innerHTML = `
                <div class="text-asset">
                    <h3>Additional Information</h3>
                    <pre>${challenge.text_asset}</pre>
                </div>
            `;
        } else {
            textAssetContainer.innerHTML = '';
        }
        
        // Show file asset if available
        const filesContainer = document.getElementById('challengeFiles');
        if (challenge.file_asset) {
            // Extract filename from path and strip parameters
            const filename = (challenge.file_asset.split('/').pop() || 'Download File').split('?')[0];
            filesContainer.innerHTML = `
                <div class="file-asset">
                    <h3>Challenge Files</h3>
                    <a href="${challenge.file_asset}" class="download-link" download>
                        📁 ${filename}
                    </a>
                </div>
            `;
        } else {
            filesContainer.innerHTML = '';
        }
    }

    // Submit flag
    async submitFlag() {
        const flagInput = document.getElementById('flagInput');
        const flag = flagInput.value.trim();
        
        if (!flag) {
            this.showFlagResult('Please enter a flag', 'error');
            return;
        }
        
        if (!this.currentChallenge) {
            this.showFlagResult('No challenge selected', 'error');
            return;
        }
        
        // For exam challenges, require confirmation before submission
        if (this.currentChallenge.is_exam) {
            const confirmed = confirm(
                'WARNING: This will cost 1 token regardless of whether your answer is correct.\n\n' +
                'Are you sure you want to submit this flag for the exam challenge?'
            );
            if (!confirmed) {
                return; // User cancelled, don't submit
            }
        }
        
        try {
            const apiEndpoint = this.currentChallenge.is_exam 
                ? `/api/adoble/${this.currentChallenge.id}/submission`
                : `/api/challenges/${this.currentChallenge.id}/submission`;
                
            const response = await this.apiCall(apiEndpoint, {
                method: 'POST',
                body: JSON.stringify({
                    flag: flag
                })
            });
            
            const result = await response.json();
            
            if (response.ok) {
                if (result.message === 'Challenge completed successfully!' || result.message === 'Exam challenge completed successfully!') {
                    const isExam = this.currentChallenge.is_exam;
                    let successMessage;
                    
                    if (isExam) {
                        successMessage = `Correct! You earned 1 token (cost: 1 token burned)!`;
                    } else {
                        successMessage = `Correct! You earned 1 token and ${result.points_earned} points!`;
                    }
                    
                    this.showFlagResult(successMessage, 'success');
                    
                    // Update current challenge object to mark as completed
                    this.currentChallenge.completed = true;
                    
                    // Update challenge status in challenges array (only for regular challenges)
                    if (!isExam) {
                        this.updateChallengeStatus(this.currentChallenge.id, 'solved', true);
                    }
                    
                    // Refresh user profile to get updated token count
                    await this.loadUserProfile();
                    this.showUserInfo();
                    
                    // Update challenge status
                    const statusElement = document.getElementById('challengeStatus');
                    statusElement.textContent = 'Solved ★';
                    statusElement.className = 'status completed';
                    
                    flagInput.value = '';
                } else if (result.message === 'Challenge already completed') {
                    this.currentChallenge.completed = true;
                    if (!this.currentChallenge.is_exam) {
                        this.updateChallengeStatus(this.currentChallenge.id, 'solved', true);
                    }
                    this.showFlagResult('Challenge already completed', 'error');
                } else {
                    let errorMessage = result.message || 'Incorrect flag';
                    if (result.tokens_burned) {
                        errorMessage += ` (1 token burned)`;
                        // Refresh user profile to reflect burned token for exam challenges
                        await this.loadUserProfile();
                        this.showUserInfo();
                    }
                    this.showFlagResult(errorMessage, 'error');
                }
            } else {
                const errorMsg = result.error || result.message || `HTTP ${response.status}: Error submitting flag`;
                this.showFlagResult(errorMsg, 'error');
            }
        } catch (error) {
            console.error('Error submitting flag:', error);
            this.showFlagResult(`Network error: ${error.message}`, 'error');
        }
    }

    // Show flag submission result
    showFlagResult(message, type) {
        const resultDiv = document.getElementById('flagResult');
        resultDiv.textContent = message;
        resultDiv.className = `flag-result ${type}`;
        resultDiv.style.display = 'block'; // Show the box
    }

    // Show error message
    showError(message) {
        const container = document.getElementById('challengesList');
        container.innerHTML = `<div class="auth-error">${message}</div>`;
        this.showChallengesList();
    }

    // Go back to home page
    goBackToHome() {
        this.updateURL();
        // Re-render challenges to reflect any status changes
        this.renderChallengesFromExisting();
        this.currentChallenge = null;
    }

    // Go back to exam list
    goBackToExamList() {
        this.showExamChallenges();
        this.currentChallenge = null;
    }

    // Render challenges from existing data without re-grouping
    renderChallengesFromExisting() {
        const container = document.getElementById('challengesList');
        container.innerHTML = '';
        
        this.showChallengesList();

        this.challenges.forEach(category => {
            const categoryDiv = document.createElement('div');
            categoryDiv.className = 'challenge-category';
            
            const categoryTitle = document.createElement('h2');
            categoryTitle.textContent = this.prettifyCategory(category.category);
            categoryDiv.appendChild(categoryTitle);
            
            const gridDiv = document.createElement('div');
            gridDiv.className = 'challenge-grid';
            
            category.challenges.forEach(challenge => {
                const cardDiv = document.createElement('div');
                cardDiv.className = 'challenge-card';
                
                // Add status classes - solved takes precedence over accessed
                if (challenge.solved) {
                    cardDiv.classList.add('completed');
                } else if (challenge.accessed) {
                    cardDiv.classList.add('accessed');
                }
                
                const statusText = challenge.solved ? '★ Solved' : 
                                  (challenge.accessed ? '◐ Viewed' : '');
                const statusClass = challenge.solved ? 'status-solved' : 
                                   (challenge.accessed ? 'status-accessed' : '');
                
                cardDiv.innerHTML = `
                    <div class="challenge-name">${challenge.name}</div>
                    <div class="challenge-description">${challenge.description}</div>
                    <div class="challenge-meta">
                        <span class="challenge-points">${challenge.point_reward_amount} ${challenge.point_reward_amount === 1 ? 'point' : 'points'}</span>
                        ${statusText ? `<span class="challenge-status ${statusClass}">${statusText}</span>` : ''}
                    </div>
                `;
                
                cardDiv.addEventListener('click', () => this.showChallenge(challenge));
                gridDiv.appendChild(cardDiv);
            });
            
            categoryDiv.appendChild(gridDiv);
            container.appendChild(categoryDiv);
        });
    }

    // Bind event listeners
    bindEvents() {
        document.getElementById('backBtn').addEventListener('click', () => {
            if (this.currentChallenge && this.currentChallenge.is_exam) {
                this.goBackToExamList();
            } else {
                this.goBackToHome();
            }
        });
        
        document.getElementById('submitFlag').addEventListener('click', () => {
            this.submitFlag();
        });
        
        document.getElementById('flagInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.submitFlag();
            }
        });
        
        document.getElementById('goToExamBtn').addEventListener('click', () => {
            this.showExamChallenges();
        });
        
        document.getElementById('examBackBtn').addEventListener('click', () => {
            this.goBackToHome();
        });
        
        // Alias management event listeners
        document.getElementById('username').addEventListener('click', () => {
            this.promptSetAlias();
        });
        
        document.getElementById('userAliasText').addEventListener('click', () => {
            this.promptSetAlias();
        });
        
        document.getElementById('removeAlias').addEventListener('click', (e) => {
            e.stopPropagation(); // Prevent triggering the alias text click
            this.removeAlias();
        });
        
        // Handle browser back/forward navigation
        window.addEventListener('popstate', () => {
            this.handleRouting();
            const urlParams = new URLSearchParams(window.location.search);
            const challengeId = urlParams.get('challenge');
            const examView = urlParams.get('exam');
            
            if (examView === 'true') {
                this.showExamChallenges();
            } else if (challengeId && this.challenges.length > 0) {
                const challenge = this.findChallengeById(challengeId);
                if (challenge) {
                    this.showChallenge(challenge);
                } else {
                    this.goBackToHome();
                }
            } else {
                this.goBackToHome();
            }
        });
    }

    // Show exam challenges page
    async showExamChallenges() {
        this.updateURL(null, true);
        document.getElementById('challengesList').style.display = 'none';
        document.getElementById('infoSection').style.display = 'none';
        document.getElementById('flagNote').style.display = 'none';
        document.getElementById('examSection').style.display = 'none';
        document.getElementById('challengeDetail').style.display = 'none';
        document.getElementById('authError').style.display = 'none';
        document.getElementById('examList').style.display = 'block';
        
        // Update exam welcome message with user name
        if (this.userInfo) {
            const username = this.userInfo.user_email || this.userInfo.username;
            if (username) {
                document.getElementById('examWelcome').textContent = `Welcome CSIS Agent ${username}`;
            }
        }
        
        try {
            await this.loadExamChallenges();
        } catch (error) {
            console.error('Failed to load exam challenges:', error);
            this.showError(`Failed to load exam challenges: ${error.message}`);
        }
    }

    // Load exam challenges from API
    async loadExamChallenges() {
        try {
            const response = await this.apiCall('/api/adoble');
            if (response.ok) {
                const examChallenges = await response.json();
                this.renderExamChallenges(examChallenges);
            } else {
                const errorData = await response.json().catch(() => ({}));
                const errorMessage = errorData.error || errorData.message || `HTTP ${response.status}: Failed to load exam challenges`;
                throw new Error(errorMessage);
            }
        } catch (error) {
            throw error;
        }
    }

    // Render exam challenges list
    renderExamChallenges(examChallenges) {
        const container = document.getElementById('examList');
        
        // Clear existing content except back button, header and story
        const backBtn = container.querySelector('#examBackBtn');
        const header = container.querySelector('h2');
        const story = container.querySelector('.exam-story');
        container.innerHTML = '';
        container.appendChild(backBtn);
        container.appendChild(header);
        container.appendChild(story);

        if (examChallenges.length === 0) {
            const noExamsDiv = document.createElement('div');
            noExamsDiv.className = 'auth-error';
            noExamsDiv.innerHTML = '<p>No exam challenges available yet.</p>';
            container.appendChild(noExamsDiv);
            return;
        }

        const gridDiv = document.createElement('div');
        gridDiv.className = 'challenge-grid';
        
        examChallenges.forEach(challenge => {
            const cardDiv = document.createElement('div');
            cardDiv.className = 'challenge-card';
            
            // Add status classes
            if (challenge.completed) {
                cardDiv.classList.add('completed');
            }
            
            const statusText = challenge.completed ? '★ Solved' : '';
            const statusClass = challenge.completed ? 'status-solved' : '';
            
            cardDiv.innerHTML = `
                <div class="challenge-name">${challenge.name}</div>
                <div class="challenge-description">${challenge.description}</div>
                <div class="challenge-meta">
                    <span class="challenge-points">1 token cost</span>
                    ${statusText ? `<span class="challenge-status ${statusClass}">${statusText}</span>` : ''}
                </div>
            `;
            
            cardDiv.addEventListener('click', () => this.showExamChallenge(challenge));
            gridDiv.appendChild(cardDiv);
        });
        
        container.appendChild(gridDiv);
    }

    // Show individual exam challenge detail
    async showExamChallenge(challenge) {
        this.currentChallenge = challenge;
        this.currentChallenge.is_exam = true;
        
        // Update URL to include exam challenge ID
        this.updateURL(challenge.id, false, true);
        
        // Clear previous challenge information immediately to prevent flashing
        this.clearChallengeDetails();
        
        document.getElementById('examList').style.display = 'none';
        document.getElementById('infoSection').style.display = 'none';
        document.getElementById('examSection').style.display = 'none';
        document.getElementById('challengeDetail').style.display = 'block';
        
        // Load detailed exam challenge info
        await this.loadExamChallengeDetails(challenge.id);
    }

    // Load detailed exam challenge information
    async loadExamChallengeDetails(challengeId) {
        try {
            const response = await this.apiCall(`/api/adoble/${challengeId}`);
            if (response.ok) {
                const detailedChallenge = await response.json();
                detailedChallenge.is_exam = true;
                this.renderChallengeDetails(detailedChallenge);
            } else {
                const errorData = await response.json().catch(() => ({}));
                const errorMessage = errorData.error || errorData.message || `HTTP ${response.status}: Failed to load exam challenge details`;
                throw new Error(errorMessage);
            }
        } catch (error) {
            console.error('Error loading exam challenge details:', error);
            this.showError(`Failed to load exam challenge details: ${error.message}`);
        }
    }

    // Update URL without page reload
    updateURL(challengeId = null, examView = false, examChallenge = false) {
        const url = new URL(window.location);
        url.searchParams.delete('challenge');
        url.searchParams.delete('exam');
        
        if (examView) {
            url.searchParams.set('exam', 'true');
        } else if (examChallenge) {
            url.searchParams.set('challenge', challengeId);
            url.searchParams.set('exam', 'challenge');
        } else if (challengeId) {
            url.searchParams.set('challenge', challengeId);
        }
        
        window.history.pushState({}, '', url);
    }

    // Prompt user to set or update alias
    async promptSetAlias() {
        const currentAlias = this.userInfo.alias || '';
        const message = currentAlias ? 
            `Current alias: ${currentAlias}\n\nYou can only change your alias once very 24h.\nEnter new alias (or leave empty to cancel):` :
            'You can only change your alias once every 24h.\n\nEnter an alias (or leave empty to cancel):';
            
        const newAlias = prompt(message, currentAlias);
        
        if (newAlias === null) {
            return; // User cancelled
        }
        
        const trimmedAlias = newAlias.trim();
        if (trimmedAlias === '') {
            return; // User entered empty string, treat as cancel
        }
        
        if (trimmedAlias === currentAlias) {
            return; // No change
        }
        
        try {
            const response = await this.apiCall('/api/alias', {
                method: 'POST',
                body: JSON.stringify({
                    alias: trimmedAlias
                })
            });
            
            if (response.ok) {
                // Update local user info immediately
                this.userInfo.alias = trimmedAlias;
                this.showUserInfo();
                
                // Show success message briefly
                this.showTemporaryMessage('Alias set successfully!', 'success');
            } else {
                const errorData = await response.json().catch(() => ({}));
                const errorMessage = errorData.error || errorData.message || 'Failed to set alias';
                alert(`Error: ${errorMessage}`);
            }
        } catch (error) {
            console.error('Error setting alias:', error);
            alert(`Network error: ${error.message}`);
        }
    }

    // Remove user alias
    async removeAlias() {
        if (!confirm('Are you sure you want to remove your alias? You can only change it once every 24h.')) {
            return;
        }
        
        try {
            const response = await this.apiCall('/api/alias', {
                method: 'DELETE'
            });
            
            if (response.ok) {
                // Update local user info immediately
                this.userInfo.alias = '';
                this.showUserInfo();
                
                // Show success message briefly
                this.showTemporaryMessage('Alias removed successfully!', 'success');
            } else {
                const errorData = await response.json().catch(() => ({}));
                const errorMessage = errorData.error || errorData.message || 'Failed to remove alias';
                alert(`Error: ${errorMessage}`);
            }
        } catch (error) {
            console.error('Error removing alias:', error);
            alert(`Network error: ${error.message}`);
        }
    }

    // Show temporary message
    showTemporaryMessage(message, type) {
        // Create temporary message element
        const messageDiv = document.createElement('div');
        messageDiv.className = `temporary-message ${type}`;
        messageDiv.textContent = message;
        messageDiv.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 1rem;
            border-radius: 8px;
            z-index: 1000;
            background-color: ${type === 'success' ? '#10b981' : '#ef4444'};
            color: white;
            font-weight: 500;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        `;
        
        document.body.appendChild(messageDiv);
        
        // Remove after 3 seconds
        setTimeout(() => {
            if (messageDiv.parentNode) {
                messageDiv.parentNode.removeChild(messageDiv);
            }
        }, 3000);
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new CTFApp();
});
