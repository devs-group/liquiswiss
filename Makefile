.PHONY: frontend-run backend-run

frontend-run:
	cd frontend && . $$HOME/.nvm/nvm.sh && nvm use && npm run dev

backend-run:
	cd backend && air
